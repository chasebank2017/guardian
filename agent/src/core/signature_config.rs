use serde::{Deserialize, Serialize};

/// 微信特征码配置结构体，由后端下发，支持多版本热更新
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SignatureConfig {
    /// 微信版本号，如 "4.0.0.26"
    pub wechat_version: String,
    /// 关键特征码（如偏移量、magic bytes、搜索模式等），可根据实际需求扩展
    pub magic_offsets: Vec<MagicOffset>,
    /// 其他可扩展字段
    // pub ...
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MagicOffset {
    /// 特征码名称（如 "DBKeyPattern"、"SessionKeyOffset" 等）
    pub name: String,
    /// 偏移量或特征码值
    pub value: String,
    /// 可选：描述或备注
    pub description: Option<String>,
}
