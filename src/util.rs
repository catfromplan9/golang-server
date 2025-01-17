use rand::Rng;
use std::cmp;
use std::fs;

pub fn random_string(n: usize) -> String {
    const LETTER_BYTES: &[u8] = b"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
    let mut rng = rand::thread_rng();
    (0..n)
        .map(|_| LETTER_BYTES[rng.gen_range(0..LETTER_BYTES.len())] as char)
        .collect()
}

pub fn min(a: i32, b: i32) -> i32 {
    cmp::min(a, b)
}

pub fn max(a: i32, b: i32) -> i32 {
    cmp::max(a, b)
}

pub fn exists(path: &str) -> bool {
    fs::metadata(path).is_ok()
}
