use chrono::Utc;
use sqlx::PgPool;

pub async fn add_password_reset_code_to_database(
    pool: &PgPool,
    email: &str,
    code: &str,
) -> Result<(), sqlx::Error> {
    sqlx::query(
        "UPDATE users SET reset_password_code=$1, last_send=$2, reset_password_attempts=5 WHERE email=$3",
    )
    .bind(code)
    .bind(Utc::now())
    .bind(email)
    .execute(pool)
    .await?;

    Ok(())
}

pub async fn check_reset_password_code(
    pool: &PgPool,
    email: &str,
    code: &str,
) -> Result<bool, sqlx::Error> {
    let row: Option<(i32,)> =
        sqlx::query_as("SELECT id FROM users WHERE email=$1 AND reset_password_code=$2")
            .bind(email)
            .bind(code)
            .fetch_optional(pool)
            .await?;

    Ok(row.is_some())
}

pub async fn is_code_fresh(pool: &PgPool, email: &str) -> Result<bool, sqlx::Error> {
    let row: (bool,) = sqlx::query_as(
        r#"
        SELECT EXISTS(
            SELECT 1
            FROM users
            WHERE email = $1
              AND NOW() < last_send + INTERVAL '15 minutes'
        )
        "#,
    )
    .bind(email)
    .fetch_one(pool)
    .await?;

    Ok(row.0)
}

pub async fn decrease_reset_password_attempts(
    pool: &PgPool,
    email: &str,
) -> Result<i32, sqlx::Error> {
    let row: (i32,) = sqlx::query_as(
        r#"
        UPDATE users
        SET reset_password_attempts = GREATEST(reset_password_attempts - 1, 0)
        WHERE email = $1
        RETURNING reset_password_attempts
        "#,
    )
    .bind(email)
    .fetch_one(pool)
    .await?;

    Ok(row.0)
}

pub async fn reset_password(pool: &PgPool, email: &str, password: &str) -> Result<(), sqlx::Error> {
    sqlx::query("UPDATE users SET password=$1 WHERE email=$2")
        .bind(password)
        .bind(email)
        .execute(pool)
        .await?;

    Ok(())
}
