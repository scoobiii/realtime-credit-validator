package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/redis/go-redis/v9"
    "github.com/segmentio/kafka-go"

    "github.com/scoobiii/realtime-credit-validator/src/policy"
)

func main() {
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
    if err := rdb.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("Redis connection failed: %v", err)
    }

    kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
    if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
        kafkaBrokers = []string{"localhost:9092"}
    }
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  kafkaBrokers,
        Topic:    "credits",
        GroupID:  "policy-group",
        MinBytes: 10e3,
        MaxBytes: 10e6,
    })
    defer reader.Close()

    updater := policy.NewEntitlementUpdater(rdb, reader)

    ctx, cancel := context.WithCancel(context.Background())
    go updater.Start(ctx)

    log.Println("Policy service running. Waiting for credit events...")

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    log.Println("Shutting down policy service...")
    cancel()
}
