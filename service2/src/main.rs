use std::{
    net::{Ipv4Addr, SocketAddr},
    time::Duration,
};

use axum::{
    extract::{ConnectInfo, State},
    routing, Router,
};
use futures_lite::stream::StreamExt;

type StdError = Box<dyn std::error::Error + Send + Sync + 'static>;

#[tokio::main]
async fn main() -> Result<(), StdError> {
    // Wait 2 seconds before starting.
    tokio::time::sleep(Duration::from_secs(2)).await;

    let rabbit_addr = std::env::var("RABBITMQ_ADDR").unwrap_or_default();

    let conn = loop {
        match lapin::Connection::connect(
            &rabbit_addr,
            lapin::ConnectionProperties::default().with_connection_name("service2".into()),
        )
        .await
        {
            Ok(conn) => break conn,
            Err(_) => {
                eprintln!("failed to connect; retrying in 2 seconds");
                tokio::time::sleep(Duration::from_secs(2)).await
            }
        }
    };

    let log_channel = conn.create_channel().await?;
    log_channel
        .queue_declare(
            "log",
            lapin::options::QueueDeclareOptions::default(),
            lapin::types::FieldTable::default(),
        )
        .await?;

    let msg_channel = conn.create_channel().await?;
    msg_channel
        .queue_declare(
            "message",
            lapin::options::QueueDeclareOptions::default(),
            lapin::types::FieldTable::default(),
        )
        .await?;

    msg_channel
        .queue_bind(
            "message",
            "message",
            "#",
            lapin::options::QueueBindOptions::default(),
            lapin::types::FieldTable::default(),
        )
        .await?;

    let mut consumer = msg_channel
        .basic_consume(
            "message",
            "service2",
            lapin::options::BasicConsumeOptions::default(),
            lapin::types::FieldTable::default(),
        )
        .await?;

    tokio::spawn({
        let log_channel = log_channel.clone();
        async move {
            eprintln!("here we are");

            while let Some(delivery) = consumer.next().await {
                let delivery = delivery.expect("error in consumer");
                delivery
                    .ack(lapin::options::BasicAckOptions::default())
                    .await
                    .expect("failed to ack");

                eprintln!("service2 got delivery");

                let delivery_data = String::from_utf8_lossy(&delivery.data);
                let log_line = format!("{delivery_data} MSG");

                log_channel
                    .basic_publish(
                        "log",
                        "",
                        lapin::options::BasicPublishOptions::default(),
                        log_line.as_bytes(),
                        lapin::BasicProperties::default(),
                    )
                    .await
                    .expect("failed to publish")
                    .await
                    .expect("failed to publish");
            }
        }
    });

    // Create router with single POST handler and select socket address.
    let router = Router::new().route("/", routing::post(request_handler).with_state(log_channel));
    let socket_address = SocketAddr::from((Ipv4Addr::UNSPECIFIED, 8000));

    // Bind and start the server.
    axum::Server::bind(&socket_address)
        .serve(router.into_make_service_with_connect_info::<SocketAddr>())
        .with_graceful_shutdown(shutdown_signal())
        .await
        .expect("failed to start or run the service");

    Ok(())
}

/// Handler of incoming POST requests at root address.
async fn request_handler(
    State(channel): State<lapin::Channel>,
    ConnectInfo(addr): ConnectInfo<SocketAddr>,
    data: String,
) {
    let log_line = format!("{data} {addr}");

    channel
        .basic_publish(
            "log",
            "",
            lapin::options::BasicPublishOptions::default(),
            log_line.as_bytes(),
            lapin::BasicProperties::default(),
        )
        .await
        .expect("failed to publish")
        .await
        .expect("failed to publish");
}

/// Enables graceful shutdown on Ctrl+C, or termination signal.
async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c()
            .await
            .expect("failed to listen for Ctrl+C");
    };

    #[cfg(unix)]
    let terminate = async {
        use tokio::signal::unix;

        unix::signal(unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    // Wait concurrently for any termination signal.
    tokio::select! {
        _ = ctrl_c => (),
        _ = terminate => (),
    };
}
