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
use tokio::sync::oneshot;

type StdError = Box<dyn std::error::Error + Send + Sync + 'static>;

#[tokio::main]
async fn main() -> Result<(), StdError> {
    // Create truncated log file.
    fs::create_dir_all("logs")?;
    let log_file = fs::File::create("logs/service2.log")?;

    // Create channel for shutting down the server.
    let (shutdown_tx, shutdown_rx) = oneshot::channel::<()>();

    let shared_state = Arc::new(Mutex::new((log_file, Some(shutdown_tx))));
    let router = Router::new().route("/", routing::post(handler).with_state(shared_state));

    // Wait 2 seconds before starting the server.
    tokio::time::sleep(Duration::from_secs(2)).await;

    axum::Server::bind(&"0.0.0.0:8000".parse().unwrap())
        .serve(router.into_make_service_with_connect_info::<SocketAddr>())
        .with_graceful_shutdown(async {
            let _ = shutdown_rx.await;
        })
        .await
        .expect("starting server should succeed");

    Ok(())
}

async fn handler(
    State(state): State<Arc<Mutex<(fs::File, Option<oneshot::Sender<()>>)>>>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    data: String,
) {
    let mut guard = state.lock().unwrap();

    if data == "STOP" {
        let shutdown_tx = &mut guard.1;

        if let Some(shutdown_tx) = shutdown_tx.take() {
            shutdown_tx
                .send(())
                .expect("oneshot channel should be succeed");
        }
    } else {
        let file = &mut guard.0;

        let log_line = format!("{data} {addr}\n");
        file.write_all(log_line.as_bytes()).unwrap();
    }
}
