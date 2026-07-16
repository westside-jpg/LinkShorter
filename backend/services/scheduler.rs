use sqlx::PgPool;

pub async fn schedule_deleting_not_verified_users(
    pool: &PgPool,
) -> Result<Vec<String>, sqlx::Error> {
    let rows: Vec<(String,)> = sqlx::query_as(
        r#"
        SELECT email FROM users
        WHERE is_verified = FALSE
        AND created_at < NOW() - INTERVAL '3 days'
        "#,
    )
    .fetch_all(pool)
    .await?;

    let emails: Vec<String> = rows.into_iter().map(|(e,)| e).collect();

    if emails.is_empty() {
        return Ok(Vec::new());
    }

    sqlx::query(
        r#"
        DELETE FROM links
        WHERE user_id IN (
            SELECT id
            FROM users
            WHERE is_verified = FALSE
            AND created_at < NOW() - INTERVAL '3 days'
        )
        "#,
    )
    .execute(pool)
    .await?;

    sqlx::query(
        r#"
        DELETE FROM users
        WHERE is_verified = FALSE
        AND created_at < NOW() - INTERVAL '3 days'
        "#,
    )
    .execute(pool)
    .await?;

    Ok(emails)
}
