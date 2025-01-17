use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use std::collections::HashMap;
use std::fs;
use std::path::Path;
use std::sync::{Arc, Mutex};
use std::process::Command;

#[derive(Clone)]
struct TemplateServer {
    cache: Arc<Mutex<HashMap<String, tera::Tera>>>,
}

impl TemplateServer {
    fn new() -> Self {
        Self {
            cache: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    async fn clear(&self) -> impl Responder {
        let mut cache = self.cache.lock().unwrap();
        cache.clear();
        HttpResponse::Ok().body("Cache cleared")
    }

    fn create(&self, fp: &str) -> Result<(), String> {
        let mut cache = self.cache.lock().unwrap();
        if !cache.contains_key(fp) {
            if Path::new(fp).extension().and_then(|s| s.to_str()) == Some("md") {
                let output = Command::new("node")
                    .arg("./markdown/convert.mjs")
                    .arg(fp)
                    .output()
                    .map_err(|e| e.to_string())?;
                if !output.status.success() {
                    return Err("Markdown conversion failed".into());
                }

                let buffer = fs::read_to_string("buffer").map_err(|e| e.to_string())?;
                let mut tera = tera::Tera::default();
                tera.add_raw_template(Path::new(fp).file_name().unwrap().to_str().unwrap(), &buffer)
                    .map_err(|e| e.to_string())?;
                cache.insert(fp.to_string(), tera);
            } else {
                let mut tera = tera::Tera::default();
		tera.add_template_files(vec![
		    ("./html/template.html", Some("")),
		    (fp, Some("")),
		])
		.map_err(|e| e.to_string())?;
                cache.insert(fp.to_string(), tera);
            }
        }
        Ok(())
    }

    async fn serve(&self, req_path: web::Path<String>) -> impl Responder {
        let fp = format!("./html/{}", req_path.into_inner());
        if Path::new(&fp).is_dir() {
            let redirect_path = format!("{}/", fp.trim_end_matches('/'));
            return HttpResponse::SeeOther()
                .append_header(("Location", redirect_path))
                .finish();
        }

        if let Err(e) = self.create(&fp) {
            return HttpResponse::InternalServerError().body(e);
        }

        let cache = self.cache.lock().unwrap();
        if let Some(tera) = cache.get(&fp) {
            let rendered = tera.render(Path::new(&fp).file_name().unwrap().to_str().unwrap(), &tera::Context::new());
            match rendered {
                Ok(content) => HttpResponse::Ok().body(content),
                Err(_) => HttpResponse::InternalServerError().body("Template rendering failed"),
            }
        } else {
            HttpResponse::NotFound().body("Template not found")
        }
    }
}

async fn handler_stats() -> impl Responder {
    let mut response = String::new();

    let df = Command::new("df").output();
    if let Ok(output) = df {
        response.push_str("# df\n");
        response.push_str(&String::from_utf8_lossy(&output.stdout));
    }

    let free = Command::new("free").output();
    if let Ok(output) = free {
        response.push_str("# free\n");
        response.push_str(&String::from_utf8_lossy(&output.stdout));
    }

    let uptime = Command::new("uptime").output();
    if let Ok(output) = uptime {
        response.push_str("# uptime\n");
        response.push_str(&String::from_utf8_lossy(&output.stdout));
    }

    HttpResponse::Ok().body(response)
}

async fn handler_teapot() -> impl Responder {
    HttpResponse::build(actix_web::http::StatusCode::IM_A_TEAPOT).finish()
}

async fn handler_headers(req: actix_web::HttpRequest) -> impl Responder {
    let mut response = String::new();
    for (key, value) in req.headers() {
        response.push_str(&format!("{}: {}\n", key, value.to_str().unwrap_or("")));
    }
    HttpResponse::Ok().body(response)
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let ts = TemplateServer::new();

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(ts.clone()))
            .route("/_stats", web::get().to(handler_stats))
            .route("/teapot", web::get().to(handler_teapot))
            .route("/_headers", web::get().to(handler_headers))
            .route("/_clear", web::get().to(clear_handler))
            .route("/{path:.*}", web::get().to(serve_handler))

    })
    .bind("0.0.0.0:5001")?
    .run()
    .await
}

async fn clear_handler(ts: web::Data<TemplateServer>) -> impl Responder {
    ts.clear();
    HttpResponse::Ok().finish()
}

async fn serve_handler(ts: web::Data<TemplateServer>, path: web::Path<String>) -> impl Responder {
    ts.serve(path);
    HttpResponse::Ok().finish()
}
