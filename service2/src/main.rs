use std::{
    fs,
    io::Write,
    net::SocketAddr,
    sync::{Arc, Mutex},
    time::Duration,
};

use axum::{
    extract::{ConnectInfo, State},
    routing, Router,
};

type StdError = Box<dyn std::error::Error + Send + Sync + 'static>;

#[tokio::main]
async fn main() -> Result<(), StdError> {
    // Create log file.
    fs::create_dir_all("logs")?;
    let log_file = fs::File::create("logs/service2.log")?;
    let log_file = Arc::new(Mutex::new(log_file));

    let router = Router::new().route("/", routing::post(handler).with_state(log_file));

    // Wait 2 seconds before starting the server.
    tokio::time::sleep(Duration::from_secs(2)).await;

    axum::Server::bind(&"0.0.0.0:8000".parse().unwrap())
        .serve(router.into_make_service_with_connect_info::<SocketAddr>())
        .await
        .expect("starting server should succeed");

    Ok(())
}

async fn handler(
    State(state): State<Arc<Mutex<fs::File>>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    data: String,
) {
    let log_line = format!("{data} {addr}\n");

    let mut file = state.lock().unwrap();
    file.write_all(log_line.as_bytes()).unwrap();
}
