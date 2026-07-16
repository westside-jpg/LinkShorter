use sqlx::PgPool;

use crate::models::Link;
use crate::services::utils::to_upper_and_lower;

const BASE_URL: &str = "localhost:8080/";

fn generate_random_code() -> String {
    let uuid = uuid::Uuid::new_v4().to_string();
    to_upper_and_lower(&uuid)[0..6].to_string()
}

pub async fn generate_link(
    pool: &PgPool,
    long_link: &str,
    user_id: i32,
) -> Result<String, sqlx::Error> {
    if user_id == 0 {
        let existing: Option<(String,)> = sqlx::query_as(
            "SELECT short_url FROM links WHERE original_url = $1 AND user_id = 0",
        )
        .bind(long_link)
        .fetch_optional(pool)
        .await?;

        if let Some((short_url,)) = existing {
            return Ok(short_url);
        }

        loop {
            let candidate = format!("{}{}", BASE_URL, generate_random_code());

            let taken: Option<(i32,)> =
                sqlx::query_as("SELECT id FROM links WHERE short_url = $1")
                    .bind(&candidate)
                    .fetch_optional(pool)
                    .await?;

            if taken.is_none() {
                sqlx::query(
                    "INSERT INTO links (original_url, short_url) VALUES ($1, $2)",
                )
                .bind(long_link)
                .bind(&candidate)
                .execute(pool)
                .await?;

                return Ok(candidate);
            }
        }
    }

    let existing: Option<(String,)> = sqlx::query_as(
        "SELECT short_url FROM links WHERE original_url = $1 AND user_id = $2 AND is_custom = FALSE",
    )
    .bind(long_link)
    .bind(user_id)
    .fetch_optional(pool)
    .await?;

    if let Some((short_url,)) = existing {
        return Ok(short_url);
    }

    loop {
        let candidate = format!("{}{}", BASE_URL, generate_random_code());

        let taken: Option<(i32,)> = sqlx::query_as("SELECT id FROM links WHERE short_url = $1")
            .bind(&candidate)
            .fetch_optional(pool)
            .await?;

        if taken.is_none() {
            sqlx::query(
                "INSERT INTO links (original_url, short_url, user_id) VALUES ($1, $2, $3)",
            )
            .bind(long_link)
            .bind(&candidate)
            .bind(user_id)
            .execute(pool)
            .await?;

            return Ok(candidate);
        }
    }
}

pub async fn return_original_link(pool: &PgPool, code: &str) -> Result<String, sqlx::Error> {
    let short_url = format!("{}{}", BASE_URL, code);

    let row: (String,) = sqlx::query_as("SELECT original_url FROM links WHERE short_url = $1")
        .bind(&short_url)
        .fetch_one(pool)
        .await?;

    Ok(row.0)
}

pub async fn user_links(
    pool: &PgPool,
    user_id: i32,
    search: &str,
    sort_by: &str,
    order_by: &str,
    filter_period: &str,
    filter_views: &str,
    filter_tags: &str,
) -> Result<Vec<Link>, String> {
    let search_trimmed = search.trim();
    let use_search = !search_trimmed.is_empty();

    let sort_column = match sort_by {
        "date" => "created_at",
        "views" => "views",
        "tag" => r#"LOWER(tag) COLLATE "ru-x-icu""#,
        _ => return Err("Неверная сортировка".to_string()),
    };

    let sort_order = match order_by {
        "asc" => "ASC",
        "desc" => "DESC",
        _ => return Err("Неверный порядок сортировки".to_string()),
    };

    let filter_period_query = match filter_period {
        "all" => "",
        "week" => "AND created_at >= NOW() - INTERVAL '7 days'",
        "month" => "AND created_at >= NOW() - INTERVAL '1 month'",
        "year" => "AND created_at >= NOW() - INTERVAL '1 year'",
        _ => return Err("Неверный параметр фильтрации по периоду".to_string()),
    };

    let filter_views_query = match filter_views {
        "0" => "",
        "10" => "AND views >= 10",
        "100" => "AND views >= 100",
        "1000" => "AND views >= 1000",
        "10000" => "AND views >= 10000",
        _ => return Err("Неверный параметр фильтрации по просмотрам".to_string()),
    };

    let filter_tags_query = match filter_tags {
        "all" => "",
        "with" => "AND tag <> ''",
        "without" => "AND tag = ''",
        _ => return Err("Неверный параметр фильтрации по тэгам".to_string()),
    };

    let search_query = if use_search { "AND tag ILIKE $2" } else { "" };

    let query = format!(
        r#"
        SELECT id,
               short_url,
               original_url,
               views,
               tag,
               created_at
        FROM links
        WHERE user_id = $1
        {}
        {}
        {}
        {}
        ORDER BY {} {}
        "#,
        filter_period_query,
        filter_views_query,
        filter_tags_query,
        search_query,
        sort_column,
        sort_order
    );

    let result = if use_search {
        let pattern = format!("%{}%", search_trimmed);
        sqlx::query_as::<_, Link>(&query)
            .bind(user_id)
            .bind(pattern)
            .fetch_all(pool)
            .await
    } else {
        sqlx::query_as::<_, Link>(&query)
            .bind(user_id)
            .fetch_all(pool)
            .await
    };

    result.map_err(|e| e.to_string())
}

pub async fn increase_link_views(pool: &PgPool, code: &str) -> Result<(), sqlx::Error> {
    sqlx::query("UPDATE links SET views = views + 1 WHERE short_url LIKE '%' || $1 || '%'")
        .bind(code)
        .execute(pool)
        .await?;

    Ok(())
}

pub async fn delete_link(pool: &PgPool, link_id: i32, user_id: i32) -> Result<(), sqlx::Error> {
    sqlx::query("DELETE FROM links WHERE id=$1 AND user_id=$2")
        .bind(link_id)
        .bind(user_id)
        .execute(pool)
        .await?;

    Ok(())
}

pub async fn add_tag(pool: &PgPool, link_id: i32, tag: &str) -> Result<(), sqlx::Error> {
    sqlx::query("UPDATE links SET tag=$1 WHERE id=$2")
        .bind(tag)
        .bind(link_id)
        .execute(pool)
        .await?;

    Ok(())
}

pub async fn create_custom_link(
    pool: &PgPool,
    user_id: i32,
    original_link: &str,
    custom: &str,
) -> Result<(String, bool), sqlx::Error> {
    let link = format!("{}{}", BASE_URL, custom);

    let row: (bool,) = sqlx::query_as("SELECT EXISTS(SELECT 1 FROM links WHERE short_url=$1)")
        .bind(&link)
        .fetch_one(pool)
        .await?;

    if row.0 {
        return Ok(("".to_string(), true));
    }

    sqlx::query(
        "INSERT INTO links (user_id, original_url, short_url, is_custom) VALUES ($1, $2, $3, $4)",
    )
    .bind(user_id)
    .bind(original_link)
    .bind(&link)
    .bind(true)
    .execute(pool)
    .await?;

    Ok((link, false))
}
