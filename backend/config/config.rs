use std::env;

pub fn get_database_url() -> String {
    if dotenvy::dotenv().is_err() {
        println!(".env");
    }

    env::var("DATABASE_URL").expect("DATABASE_URL не задан")
}

pub fn get_jwt_secret() -> String {
    env::var("JWT_SECRET").expect("JWT_SECRET не задан")
}
