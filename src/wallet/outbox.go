package wallet

import (
    "context"
    "database/sql"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

type OutboxPublisher struct {
    db          *sql.DB
    kafkaWriter *kafka.Writer
    batchSize   int
    interval    time.Duration
}

func NewOutboxPublisher(db *sql.DB, kw *kafka.Writer) *OutboxPublisher {
    return &OutboxPublisher{
        db:          db,
        kafkaWriter: kw,
        batchSize:   100,
        interval:    100 * time.Millisecond,
    }
}

func (p *OutboxPublisher) Run(ctx context.Context) {
    ticker := time.NewTicker(p.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := p.publishBatch(ctx); err != nil {
                log.Printf("outbox publish error: %v", err)
            }
        }
    }
}

func (p *OutboxPublisher) publishBatch(ctx context.Context) error {
    rows, err := p.db.QueryContext(ctx, `
        SELECT id, aggregate_id, event_type, payload
        FROM outbox_events
        WHERE NOT processed
        ORDER BY created_at
        LIMIT $1
    `, p.batchSize)
    if err != nil {
        return err
    }
    defer rows.Close()

    var (
        ids       []string
        messages  []kafka.Message
    )

    for rows.Next() {
        var id, aggID, evtType string
        var payload []byte
        if err := rows.Scan(&id, &aggID, &evtType, &payload); err != nil {
            return err
        }
        ids = append(ids, id)
        messages = append(messages, kafka.Message{
            Topic: "credits",
            Key:   []byte(aggID),
            Value: payload,
            Headers: []kafka.Header{
                {Key: "event_type", Value: []byte(evtType)},
            },
        })
    }

    if len(messages) == 0 {
        return nil
    }

    if err := p.kafkaWriter.WriteMessages(ctx, messages...); err != nil {
        return err
    }

    _, err = p.db.ExecContext(ctx, `UPDATE outbox_events SET processed = TRUE WHERE id = ANY($1)`, ids)
    return err
}
