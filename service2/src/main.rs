use std::{
    net::{Ipv4Addr, SocketAddr},
    time::Duration,
};

use axum::{
    extract::{ConnectInfo, State},
    routing, Router,
};
use futures_lite::stream::StreamExt;
use lapin::{
    options::{
        BasicConsumeOptions, BasicPublishOptions, ExchangeDeclareOptions, QueueBindOptions,
        QueueDeclareOptions,
    },
    types::FieldTable,
    BasicProperties, ConnectionProperties, ExchangeKind,
};

type StdError = Box<dyn std::error::Error + Send + Sync + 'static>;

#[tokio::main]
async fn main() -> Result<(), StdError> {
    // Wait 2 seconds before starting.
    tokio::time::sleep(Duration::from_secs(2)).await;

    // Retry until RabbitMQ connection is established.
    let rabbit_addr = std::env::var("RABBITMQ_ADDR").unwrap_or_default();
    let conn = loop {
        match lapin::Connection::connect(
            &rabbit_addr,
            ConnectionProperties::default().with_connection_name("service2".into()),
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

    // Create channel for log topic.
    let log_channel = conn.create_channel().await?;
    log_channel
        .exchange_declare(
            "log",
            ExchangeKind::Topic,
            ExchangeDeclareOptions {
                durable: true,
                ..ExchangeDeclareOptions::default()
            },
            FieldTable::default(),
        )
        .await?;
    log_channel
        .queue_declare("log", QueueDeclareOptions::default(), FieldTable::default())
        .await?;

    // Create channel for message topic.
    let msg_channel = conn.create_channel().await?;
    msg_channel
        .exchange_declare(
            "message",
            ExchangeKind::Topic,
            ExchangeDeclareOptions {
                durable: true,
                ..ExchangeDeclareOptions::default()
            },
            FieldTable::default(),
        )
        .await?;
    msg_channel
        .queue_declare(
            "message",
            QueueDeclareOptions::default(),
            FieldTable::default(),
        )
        .await?;

    // Bind to message queue if not already.
    msg_channel
        .queue_bind(
            "message",
            "message",
            "#",
            QueueBindOptions::default(),
            FieldTable::default(),
        )
        .await?;

    // Create message channel consumer.
    let mut consumer = msg_channel
        .basic_consume(
            "message",
            "service2",
            BasicConsumeOptions {
                no_ack: true,
                ..BasicConsumeOptions::default()
            },
            FieldTable::default(),
        )
        .await?;

    // Spawn message queue listener.
    tokio::spawn({
        let log_channel = log_channel.clone();
        async move {
            while let Some(Ok(delivery)) = consumer.next().await {
                let delivery_data = String::from_utf8_lossy(&delivery.data);
                let log_line = format!("{delivery_data} MSG");

                // Publish message from message topic to log topic.
                log_channel
                    .basic_publish(
                        "log",
                        "#",
                        BasicPublishOptions::default(),
                        log_line.as_bytes(),
                        BasicProperties::default(),
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

    // Publish HTTP post requests to log topic.
    channel
        .basic_publish(
            "log",
            "",
            BasicPublishOptions::default(),
            log_line.as_bytes(),
            BasicProperties::default(),
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
