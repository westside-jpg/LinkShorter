use chrono::{DateTime, Utc};
use serde::Serialize;

#[derive(Debug, Serialize, sqlx::FromRow)]
pub struct Link {
    #[serde(rename = "id")]
    pub link_id: i32,
    #[serde(rename = "short")]
    pub short_url: String,
    #[serde(rename = "original")]
    pub original_url: String,
    pub views: i32,
    pub tag: String,
    pub created_at: DateTime<Utc>,
}
