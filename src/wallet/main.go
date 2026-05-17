package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
    "github.com/segmentio/kafka-go"

    "github.com/scoobiii/realtime-credit-validator/src/wallet"
)

func main() {
    // Conectar ao PostgreSQL
    connStr := os.Getenv("DB_URL")
    if connStr == "" {
        connStr = "postgres://user:password@localhost:5432/wallet?sslmode=disable"
    }
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    db.SetMaxOpenConns(10)

    // Executar migrations (criação de tabelas)
    initDB(db)

    // Conectar ao Kafka
    kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
    if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
        kafkaBrokers = []string{"localhost:9092"}
    }
    kafkaWriter := wallet.NewKafkaWriter(kafkaBrokers, "credits")
    defer kafkaWriter.Close()

    _ = wallet.NewLedger(db, kafkaWriter) // ledger será usado pelo gateway

    // Servidor HTTP para health check
    r := gin.Default()
    r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

    srv := &http.Server{Addr: ":8081", Handler: r}
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down wallet service...")
    srv.Shutdown(context.Background())
}

func initDB(db *sql.DB) {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS credits (
            id VARCHAR(36) PRIMARY KEY,
            user_id VARCHAR(100) NOT NULL,
            amount BIGINT NOT NULL,
            status VARCHAR(20) NOT NULL,
            source VARCHAR(20) NOT NULL,
            created_at BIGINT NOT NULL,
            idempotency_key VARCHAR(255) UNIQUE NOT NULL
        );
        CREATE TABLE IF NOT EXISTS user_balance (
            user_id VARCHAR(100) PRIMARY KEY,
            balance BIGINT NOT NULL DEFAULT 0,
            updated_at BIGINT NOT NULL
        );
        CREATE INDEX IF NOT EXISTS idx_credits_user_id ON credits(user_id);
        CREATE INDEX IF NOT EXISTS idx_credits_created_at ON credits(created_at);
    `)
    if err != nil {
        log.Fatal("Migration failed: ", err)
    }
    log.Println("Database schema ready")
}
