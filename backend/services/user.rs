use chrono::{DateTime, Utc};
use sqlx::PgPool;

pub async fn is_link_owner(pool: &PgPool, user_id: i32, link_id: i32) -> Result<bool, sqlx::Error> {
    let row: (bool,) = sqlx::query_as(
        "SELECT EXISTS(SELECT 1 FROM links WHERE id=$1 AND user_id=$2)",
    )
    .bind(link_id)
    .bind(user_id)
    .fetch_one(pool)
    .await?;

    Ok(row.0)
}

pub async fn could_resend_email(
    pool: &PgPool,
    email: &str,
) -> Result<(bool, chrono::Duration), sqlx::Error> {
    let row: (Option<DateTime<Utc>>,) =
        sqlx::query_as("SELECT last_send FROM users WHERE email = $1")
            .bind(email)
            .fetch_one(pool)
            .await?;

    let last_send_time = match row.0 {
        Some(t) => t,
        None => return Ok((true, chrono::Duration::zero())),
    };

    let elapsed = Utc::now() - last_send_time;
    let remaining = chrono::Duration::minutes(1) - elapsed;

    if elapsed < chrono::Duration::minutes(1) {
        return Ok((false, remaining));
    }

    Ok((true, chrono::Duration::zero()))
}
