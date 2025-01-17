use rusqlite::{params, Connection, Result};
use std::fs;
use std::sync::Mutex;

lazy_static::lazy_static! {
    static ref DB: Mutex<Option<Connection>> = Mutex::new(None);
}

static SQL_PATH: &str = "./sql";
static DATABASE_VERSION: &str = "dev2";

fn db_create() -> Result<()> {
    let schema = fs::read_to_string(format!("{}/schema.sql", SQL_PATH))?;
    let conn = DB.lock().unwrap().as_ref().unwrap();
    conn.execute_batch(&schema)?;
    db_pair_store("version", DATABASE_VERSION)?;
    Ok(())
}

fn db_init() -> Result<()> {
    let conn = Connection::open(format!("{}/database.db", SQL_PATH))?;
    *DB.lock().unwrap() = Some(conn);

    let version = db_pair_load("version").or_else(|_| {
        db_create()?;
        db_pair_load("version")
    })?;

    if version != DATABASE_VERSION {
        return Err(rusqlite::Error::UserFunctionError(Box::new(
            "Bad database version",
        )));
    }

    Ok(())
}

fn db_pair_load(key: &str) -> Result<String> {
    let conn = DB.lock().unwrap().as_ref().unwrap();
    let mut stmt = conn.prepare("SELECT value FROM variable WHERE key=?1")?;
    let mut rows = stmt.query(params![key])?;
    if let Some(row) = rows.next()? {
        Ok(row.get(0)?)
    } else {
        Err(rusqlite::Error::QueryReturnedNoRows)
    }
}

fn db_pair_store(key: &str, value: &str) -> Result<()> {
    let conn = DB.lock().unwrap().as_ref().unwrap();
    conn.execute(
        "REPLACE INTO variable (key, value) VALUES (?1, ?2)",
        params![key, value],
    )?;
    Ok(())
}
