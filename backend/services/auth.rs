use axum_extra::extract::cookie::CookieJar;
use sqlx::PgPool;

use crate::jwt_service;

pub async fn is_username_in_database(pool: &PgPool, username: &str) -> Result<bool, sqlx::Error> {
    let row: Option<(String,)> =
        sqlx::query_as("SELECT username FROM users WHERE username = $1")
            .bind(username)
            .fetch_optional(pool)
            .await?;
    Ok(row.is_some())
}

pub async fn is_email_in_database(pool: &PgPool, email: &str) -> Result<bool, sqlx::Error> {
    let row: Option<(String,)> = sqlx::query_as("SELECT email FROM users WHERE email = $1")
        .bind(email)
        .fetch_optional(pool)
        .await?;
    Ok(row.is_some())
}

pub async fn is_login_valid(
    pool: &PgPool,
    login_input: &str,
    password: &str,
) -> Result<bool, sqlx::Error> {
    let row: Option<(String,)> =
        sqlx::query_as("SELECT password FROM users WHERE username = $1 OR email = $1")
            .bind(login_input)
            .fetch_optional(pool)
            .await?;

    let hashed_password = match row {
        Some((p,)) => p,
        None => return Ok(false),
    };

    match bcrypt::verify(password, &hashed_password) {
        Ok(valid) => Ok(valid),
        Err(_) => Ok(false),
    }
}

pub async fn register_user(
    pool: &PgPool,
    username: &str,
    email: &str,
    password: &str,
) -> Result<(), sqlx::Error> {
    let verification_code = uuid::Uuid::new_v4().to_string();

    sqlx::query(
        "INSERT INTO users (username, password, email, verification_code) VALUES ($1, $2, $3, $4)",
    )
    .bind(username)
    .bind(password)
    .bind(email)
    .bind(verification_code)
    .execute(pool)
    .await?;

    Ok(())
}

pub async fn get_user_id(pool: &PgPool, login_input: &str) -> Result<i32, sqlx::Error> {
    let row: (i32,) =
        sqlx::query_as("SELECT id FROM users WHERE username = $1 OR email = $1")
            .bind(login_input)
            .fetch_one(pool)
            .await?;
    Ok(row.0)
}

pub async fn verify_email(pool: &PgPool, code: &str) -> Result<(), sqlx::Error> {
    sqlx::query("UPDATE users SET is_verified = true WHERE verification_code = $1")
        .bind(code)
        .execute(pool)
        .await?;
    Ok(())
}

pub async fn get_user_id_by_code(pool: &PgPool, code: &str) -> Result<i32, sqlx::Error> {
    let row: (i32,) = sqlx::query_as("SELECT id FROM users WHERE verification_code = $1")
        .bind(code)
        .fetch_one(pool)
        .await?;
    Ok(row.0)
}

pub fn get_user_id_from_jwt(jar: &CookieJar) -> Result<i32, ()> {
    let token = jar.get("token").ok_or(())?.value();
    let claims = jwt_service::decode_auth_jwt(token).map_err(|_| ())?;
    Ok(claims.user_id)
}
