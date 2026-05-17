package main

import (
    "database/sql"
    "log"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"

    "github.com/scoobiii/realtime-credit-validator/src/gateway/auth"
    "github.com/scoobiii/realtime-credit-validator/src/gateway/handlers"
    "github.com/scoobiii/realtime-credit-validator/src/gateway/ratelimit"
    "github.com/scoobiii/realtime-credit-validator/src/wallet"
)

func main() {
    dbURL := os.Getenv("DB_URL")
    if dbURL == "" {
        dbURL = "postgres://user:password@localhost:5432/wallet?sslmode=disable"
    }
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    kafkaBrokers := []string{os.Getenv("KAFKA_BROKERS")}
    if len(kafkaBrokers) == 0 || kafkaBrokers[0] == "" {
        kafkaBrokers = []string{"localhost:9092"}
    }
    kafkaWriter := wallet.NewKafkaWriter(kafkaBrokers, "credits")
    defer kafkaWriter.Close()

    ledger := wallet.NewLedger(db, kafkaWriter)

    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "localhost:6379"
    }
    rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

    limiter := ratelimit.NewRateLimiter(rdb, 10, time.Second)

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "changeme-in-production"
    }
    jwtCfg := auth.NewJWTConfig(jwtSecret, "realtime-credit-validator", "anatel-gateway", 24*time.Hour)

    r := gin.Default()
    r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

    api := r.Group("/v1")
    api.Use(auth.JWTAuthMiddleware(jwtCfg))
    api.Use(ratelimit.RateLimitMiddleware(limiter))
    {
        handler := handlers.NewGatewayHandler(ledger)
        api.POST("/credit", handler.AddCredit)
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Gateway listening on :%s", port)
    r.Run(":" + port)
}
