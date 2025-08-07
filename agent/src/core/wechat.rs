#[derive(Debug)]
pub struct ChatMessage {
    pub content: String,
    pub timestamp: i64,
}
use std::{cell::RefCell, collections::HashMap, rc::Rc};
use windows::core::PWSTR;
use windows::Win32::Foundation::{GetLastError, HANDLE, INVALID_HANDLE_VALUE, MAX_PATH};
use windows::Win32::Storage::FileSystem::{GetFileVersionInfoExW, GetFileVersionInfoSizeExW, FILE_VER_GET_LOCALISED};
use windows::Win32::System::Diagnostics::ToolHelp::{CreateToolhelp32Snapshot, Process32FirstW, Process32NextW, PROCESSENTRY32W, TH32CS_SNAPPROCESS};
use windows::Win32::System::Threading::{QueryFullProcessImageNameW, PROCESS_ACCESS_RIGHTS, PROCESS_NAME_WIN32, PROCESS_QUERY_LIMITED_INFORMATION};
// agent/src/core/wechat.rs

use std::error::Error;
use std::ops::{Add, Sub};
use std::path::PathBuf;
use windows::Win32::{
    Foundation::CloseHandle,
    System::{
        Diagnostics::Debug::ReadProcessMemory,
        Threading::{OpenProcess, PROCESS_QUERY_INFORMATION, PROCESS_VM_READ},
    },
};
use regex::Regex;


#[derive(Debug, Clone)]
pub struct WechatInfo {
    pub pid: u32,
    pub version: String,
    pub account_name: String,
    pub nick_name: Option<String>,
    pub phone: Option<String>,
    pub data_dir: String,
    pub key: Option<String>,
}

#[derive(Debug, Clone, Eq, PartialEq)]
pub struct ProcessInformatcion {
    pub pid: u32,
    pub ppid: u32,
    pub name: String,
    pub exe: String,
    pub cmd: String,
}

#[derive(Debug, Clone)]
pub struct Process {
    pub pid: u32,
    _inner_handle: Rc<RefCell<HANDLE>>,
    _granted_access: RefCell<PROCESS_ACCESS_RIGHTS>,
}

impl Process {
    pub fn new(pid: u32) -> Self {
        Self {
            pid,
            _inner_handle: Rc::new(RefCell::new(INVALID_HANDLE_VALUE)),
            _granted_access: RefCell::new(PROCESS_QUERY_LIMITED_INFORMATION | PROCESS_VM_READ),
        }
    }

    pub unsafe fn get_proc_info(&self) -> Result<ProcessInformatcion, Box<dyn Error>> {
        // 打开进程并获取信息，简化实现
        let mut exe_len = MAX_PATH;
        let mut exe = [0u16; MAX_PATH as _];
        let _ = QueryFullProcessImageNameW(
            *self._inner_handle.borrow(),
            PROCESS_NAME_WIN32,
            PWSTR::from_raw(exe.as_mut_ptr()),
            &mut exe_len,
        );
        let exe = String::from_utf16(&exe).unwrap_or_default();
        let exe = exe.trim_matches('\u{0}').to_owned();
        let name = exe.split('\').last().unwrap_or("").to_owned();
        Ok(ProcessInformatcion {
            pid: self.pid,
            ppid: 0,
            name,
            exe,
            cmd: String::new(),
        })
    }
}

pub fn get_pid_by_name(pname: &str) -> Vec<u32> {
    let mut result = vec![];
    unsafe {
        let snapshot = CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS, 0)?;
        let mut entry: PROCESSENTRY32W = std::mem::zeroed();
        entry.dwSize = std::mem::size_of::<PROCESSENTRY32W>() as u32;
        if Process32FirstW(snapshot, &mut entry).as_bool() {
            loop {
                let exe_name = String::from_utf16_lossy(&entry.szExeFile);
                if exe_name.trim_end_matches('\u{0}') == pname {
                    result.push(entry.th32ProcessID);
                }
                if !Process32NextW(snapshot, &mut entry).as_bool() {
                    break;
                }
            }
        }
        CloseHandle(snapshot)?;
    }
    Ok(result)
}

pub fn read_number<T: Sub + Add + Ord + Default>(pid: u32, addr: usize) -> Result<T, Box<dyn Error>> {
    unsafe {
        let hprocess = OpenProcess(PROCESS_VM_READ | PROCESS_QUERY_INFORMATION, false, pid)?;
        let mut result: T = T::default();
        ReadProcessMemory(
            hprocess,
            addr as _,
            std::mem::transmute(&mut result),
            std::mem::size_of::<T>(),
            None,
        )?;
        CloseHandle(hprocess)?;
        Ok(result)
    }
}

pub fn read_string(pid: u32, addr: usize, size: usize) -> Result<String, Box<dyn Error>> {
    unsafe {
        let hprocess = OpenProcess(PROCESS_VM_READ | PROCESS_QUERY_INFORMATION, false, pid)?;
        let mut buffer = vec![0; size];
        let _ = ReadProcessMemory(hprocess, addr as _, buffer.as_mut_ptr() as _, size, None);
        CloseHandle(hprocess)?;
        let buf_str = match buffer.iter().position(|&x| x == 0) {
            Some(pos) => String::from_utf8(buffer[..pos].to_vec())?,
            None => String::from_utf8(buffer)?,
        };
        if buf_str.len() != size {
            Err(format!(
                "except {} characters, but found: {} --> {}",
                size,
                buf_str.len(),
                buf_str
            ).into())
        } else {
            Ok(buf_str)
        }
    }
}

pub fn read_bytes(pid: u32, addr: usize, size: usize) -> Result<Vec<u8>, Box<dyn Error>> {
    unsafe {
        let hprocess = OpenProcess(PROCESS_VM_READ | PROCESS_QUERY_INFORMATION, false, pid)?;
        let mut buffer = vec![0; size];
        let _ = ReadProcessMemory(hprocess, addr as _, buffer.as_mut_ptr() as _, size, None)?;
        CloseHandle(hprocess)?;
        Ok(buffer)
    }
}

pub fn get_wechat_data() -> Result<Vec<ChatMessage>, Box<dyn Error>> {
    // 1. 查找正在运行的 WeChat.exe 进程
    let pids = get_pid_by_name("WeChat.exe");
    let pid = *pids.first().ok_or("WeChat.exe process not found")?;

    // 2. 获取进程版本号
    let version = get_proc_file_version(pid).unwrap_or_else(|| "unknown".to_string());

    // 3. 初始化 WechatInfo 结构体，后续补充数据目录、key、账号名等
    let mut info = WechatInfo {
        pid,
        version,
        account_name: String::new(),
        nick_name: None,
        phone: None,
        data_dir: String::new(),
        key: None,
    };

    // 获取数据目录（示例：假设 WeChat 数据目录为 C:\Users\<User>\Documents\WeChat Files）
    // 实际应通过内存扫描或注册表等方式获取，这里仅演示填充流程
    let user_profile = std::env::var("USERPROFILE").unwrap_or_default();
    let default_data_dir = format!("{}\\Documents\\WeChat Files", user_profile);
    info.data_dir = default_data_dir;

    use std::fs;
    let mut messages = Vec::new();
    let entries = fs::read_dir(&info.data_dir)?;
    for entry in entries {
        let entry = entry?;
        if !entry.file_type()?.is_dir() { continue; }
        let name = entry.file_name();
        let account_name = name.to_string_lossy();
        if account_name == "All Users" || account_name == "Applet" || account_name == "FileStorage" { continue; }

        let mut info = info.clone();
        info.account_name = account_name.to_string();

        // key
        let key_path = format!("{}\{}\key", info.data_dir, info.account_name);
        let key = fs::read(&key_path)
            .ok()
            .and_then(|bytes| if !bytes.is_empty() { Some(hex::encode(bytes)) } else { None });
        info.key = key;

        // 数据库解密与消息提取
        let msg_db_path = format!("{}\{}\Msg\Msg.db", info.data_dir, info.account_name);
        if let (Some(key_hex), true) = (&info.key, std::path::Path::new(&msg_db_path).exists()) {
            #[cfg(feature = "sqlcipher")]
            {
                use rusqlite::Connection;
                use std::fs::File;
                use std::io::Write;
                use std::env::temp_dir;
                use std::path::PathBuf;
                use std::fs::remove_file;

                // 复制数据库到临时文件
                let mut tmp_path = temp_dir();
                tmp_path.push(format!("wechat_msg_{}_tmp.db", info.account_name));
                std::fs::copy(&msg_db_path, &tmp_path)?;

                let mut conn = Connection::open(&tmp_path)?;
                conn.pragma_update(None, "key", &key_hex)?;

                let mut stmt = conn.prepare("SELECT StrContent, CreateTime FROM MSG")?;
                let rows = stmt.query_map([], |row| {
                    let content: String = row.get(0)?;
                    let timestamp: i64 = row.get(1)?;
                    Ok(ChatMessage { content, timestamp })
                })?;
                for msg in rows.flatten() {
                    messages.push(msg);
                }
                // 删除临时文件
                let _ = remove_file(&tmp_path);
            }
        }
    }
    Ok(messages)
}
// 获取进程文件版本号
pub fn get_proc_file_version(pid: u32) -> Option<String> {
    unsafe {
        let process = Process::new(pid);
        match process.get_file_info() {
            Ok(info) => info.get("FileVersion").cloned(),
            Err(_) => None,
        }
    }
}

impl Process {
    pub unsafe fn get_file_info(&self) -> Result<HashMap<String, String>, Box<dyn Error>> {
        // 仅实现 FileVersion 获取，简化版
        let pi = self.get_proc_info()?;
        let mut temp: u32 = 0;
        let mut exe = pi.exe.encode_utf16().collect::<Vec<u16>>();
        exe.push(0x00);
        let len = GetFileVersionInfoSizeExW(FILE_VER_GET_LOCALISED, PWSTR(exe.as_mut_ptr()), &mut temp);
        if len == 0 {
            return Err("Failed to get file version info size".into());
        }
        let mut addr = vec![0u16; len as usize / 2 + 1];
        let mut hash: HashMap<String, String> = HashMap::new();
        match GetFileVersionInfoExW(
            FILE_VER_GET_LOCALISED,
            PWSTR(exe.as_mut_ptr()),
            0,
            len,
            addr.as_mut_ptr() as _,
        ) {
            Ok(_) => {
                // 这里只简单返回 FileVersion
                hash.insert("FileVersion".to_string(), "unknown".to_string());
                Ok(hash)
            },
            Err(_) => Err("Failed to get file version info".into()),
        }
    }
}
