use chrono::Utc;
use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
use serde::{Deserialize, Serialize};

use crate::config;

#[derive(Debug, Serialize, Deserialize)]
pub struct AuthClaims {
    pub user_id: i32,
    pub exp: usize,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ResetPasswordClaims {
    pub email: String,
    pub purpose: String,
    pub exp: usize,
}

pub fn generate_jwt(user_id: i32) -> Result<String, jsonwebtoken::errors::Error> {
    let exp = (Utc::now() + chrono::Duration::days(30)).timestamp() as usize;
    let claims = AuthClaims { user_id, exp };
    encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(config::get_jwt_secret().as_bytes()),
    )
}

pub fn generate_reset_password_jwt(email: &str) -> Result<String, jsonwebtoken::errors::Error> {
    let exp = (Utc::now() + chrono::Duration::minutes(15)).timestamp() as usize;
    let claims = ResetPasswordClaims {
        email: email.to_string(),
        purpose: "reset_password".to_string(),
        exp,
    };
    encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(config::get_jwt_secret().as_bytes()),
    )
}

pub fn decode_auth_jwt(token: &str) -> Result<AuthClaims, jsonwebtoken::errors::Error> {
    let data = decode::<AuthClaims>(
        token,
        &DecodingKey::from_secret(config::get_jwt_secret().as_bytes()),
        &Validation::default(),
    )?;
    Ok(data.claims)
}

pub fn decode_reset_password_jwt(
    token: &str,
) -> Result<ResetPasswordClaims, jsonwebtoken::errors::Error> {
    let data = decode::<ResetPasswordClaims>(
        token,
        &DecodingKey::from_secret(config::get_jwt_secret().as_bytes()),
        &Validation::default(),
    )?;
    Ok(data.claims)
}
