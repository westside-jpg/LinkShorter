use rand::Rng;

pub fn to_upper_and_lower(text: &str) -> String {
    let mut rng = rand::thread_rng();
    text.chars()
        .map(|c| {
            if rng.gen_range(0..2) == 0 {
                c.to_uppercase().next().unwrap_or(c)
            } else {
                c.to_lowercase().next().unwrap_or(c)
            }
        })
        .collect()
}

pub fn generate_verification_code() -> String {
    let mut rng = rand::thread_rng();
    let code: u32 = rng.gen_range(0..1_000_000);
    format!("{:06}", code)
}

pub fn declination_word(n: i64, one: &str, two: &str, many: &str) -> String {
    let last_two_digits = n % 100;

    if (11..=14).contains(&last_two_digits) {
        return many.to_string();
    }

    match n % 10 {
        1 => one.to_string(),
        2 | 3 | 4 => two.to_string(),
        _ => many.to_string(),
    }
}
