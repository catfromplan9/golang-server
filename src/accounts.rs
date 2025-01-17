use argon2::{self, Config};
use chrono::Utc;
use rusqlite::{params, Connection, Result};
use std::sync::Mutex;
use tera::{Context, Tera};
use warp::{http::Response, reject, reply, Filter, Rejection, Reply};

lazy_static::lazy_static! {
    static ref DB: Mutex<Option<Connection>> = Mutex::new(None);
    static ref TEMPLATES: Tera = Tera::new("html/**/*").unwrap();
}

pub struct Account {
    pub id: i32,
    pub username: String,
    pub email: String,
    pub hash: String,
    pub token: Option<String>,
    pub token_issued: Option<i64>,
    pub created: i64,
    pub verified: i64,
    pub class: i32,
}

fn hash_password(password: &str) -> Result<String> {
    let salt = b"somesalt";
    let config = Config::default();
    argon2::hash_encoded(password.as_bytes(), salt, &config).map_err(|_| rusqlite::Error::InvalidParameterName)
}

fn check_password(password: &str, hash: &str) -> bool {
    argon2::verify_encoded(hash, password.as_bytes()).unwrap_or(false)
}

pub fn accounts_init() -> Result<()> {
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    TEMPLATES.render("register.html", &Context::new()).unwrap();
    TEMPLATES.render("login.html", &Context::new()).unwrap();
    Ok(())
}

fn check_username(username: &str) -> Result<Option<String>> {
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    let mut stmt = conn.prepare("SELECT COUNT(*) FROM account WHERE username = ?1")?;
    let count: i32 = stmt.query_row(params![username], |row| row.get(0))?;
    if count > 0 {
        Ok(Some("Username occupied".to_string()))
    } else {
        Ok(None)
    }
}

fn check_email(email: &str) -> Result<Option<String>> {
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    let mut stmt = conn.prepare("SELECT COUNT(*) FROM account WHERE email = ?1")?;
    let count: i32 = stmt.query_row(params![email], |row| row.get(0))?;
    if count > 0 {
        Ok(Some("Email already in use".to_string()))
    } else {
        Ok(None)
    }
}

fn account_token_new(id: i32) -> Result<String> {
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    let token = crate::util::random_string(128);
    let token_issued = Utc::now().timestamp();
    conn.execute(
        "UPDATE account SET token = ?1, token_issued = ?2 WHERE id = ?3",
        params![token, token_issued, id],
    )?;
    Ok(token)
}

fn account_token_get(id: i32) -> Result<String> {
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    let mut stmt = conn.prepare("SELECT token, token_issued FROM account WHERE id = ?1")?;
    let mut rows = stmt.query(params![id])?;
    if let Some(row) = rows.next()? {
        let token: Option<String> = row.get(0)?;
        if let Some(tok) = token {
            Ok(tok)
        } else {
            account_token_new(id)
        }
    } else {
        Err(rusqlite::Error::QueryReturnedNoRows)
    }
}

fn account_register(username: &str, password: &str, email: &str) -> Result<Option<String>> {
    let hash = hash_password(password)?;
    if let Some(status) = check_username(username)? {
        return Ok(Some(status));
    }
    if let Some(status) = check_email(email)? {
        return Ok(Some(status));
    }
    let created = Utc::now().timestamp();
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    conn.execute(
        "INSERT INTO account (username, email, hash, created, verified, power) VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
        params![username, email, hash, created, 0, 0],
    )?;
    Ok(None)
}

fn account_login(username: &str, password: &str) -> Result<Option<String>> {
    let conn = DB.lock().unwrap();
    let conn = conn.as_ref().unwrap();
    let mut stmt = conn.prepare("SELECT id, hash FROM account WHERE username = ?1")?;
    let mut rows = stmt.query(params![username])?;
    while let Some(row) = rows.next()? {
        let id: i32 = row.get(0)?;
        let hash: String = row.get(1)?;
        if check_password(password, &hash) {
            return account_token_get(id).map(Some);
        }
    }
    Ok(None)
}
