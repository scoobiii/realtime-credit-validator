package policy

import (
    "context"
    "encoding/json"
    "log"

    "github.com/redis/go-redis/v9"
    "github.com/segmentio/kafka-go"
)

type CreditEvent struct {
    Event     string `json:"event"`
    UserID    string `json:"user_id"`
    Amount    int64  `json:"amount"`
    Source    string `json:"source"`
    Timestamp int64  `json:"timestamp"`
}

type EntitlementUpdater struct {
    redisClient *redis.Client
    kafkaReader *kafka.Reader
}

func NewEntitlementUpdater(redisClient *redis.Client, kafkaReader *kafka.Reader) *EntitlementUpdater {
    return &EntitlementUpdater{
        redisClient: redisClient,
        kafkaReader: kafkaReader,
    }
}

func (e *EntitlementUpdater) Start(ctx context.Context) {
    log.Println("Policy service started, waiting for credits...")
    for {
        msg, err := e.kafkaReader.ReadMessage(ctx)
        if err != nil {
            log.Printf("Error reading Kafka: %v", err)
            continue
        }

        var event CreditEvent
        if err := json.Unmarshal(msg.Value, &event); err != nil {
            log.Printf("Error parsing event: %v", err)
            continue
        }

        if event.Event != "CreditoConfirmado" {
            continue
        }

        log.Printf("Processing credit for user %s, amount %d", event.UserID, event.Amount)

        // Atualiza franquia (Redis)
        if err := e.redisClient.IncrBy(ctx, "quota:"+event.UserID, event.Amount).Err(); err != nil {
            log.Printf("Failed to update quota: %v", err)
            continue
        }

        // Publica comando para derrubar sessão
        if err := e.redisClient.Publish(ctx, "session:kill", event.UserID).Err(); err != nil {
            log.Printf("Failed to publish session kill: %v", err)
        } else {
            log.Printf("Session kill command sent for user %s", event.UserID)
        }
    }
}
