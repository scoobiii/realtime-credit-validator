package wallet\n\nfunc AddCredit() {}

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"

    "github.com/google/uuid"
    "github.com/lib/pq"
    "github.com/segmentio/kafka-go"
)

type Ledger struct {
    db          *sql.DB
    kafkaWriter *kafka.Writer
}

type CreditTransaction struct {
    ID         string `json:"id"`
    UserID     string `json:"user_id"`
    Amount     int64  `json:"amount"`
    Status     string `json:"status"`
    Source     string `json:"source"`
    CreatedAt  int64  `json:"created_at"`
}

func NewLedger(db *sql.DB, kafkaWriter *kafka.Writer) *Ledger {
    return &Ledger{db: db, kafkaWriter: kafkaWriter}
}

// AddCredit realiza a transação em uma única operação ACID e publica evento no Kafka
func (l *Ledger) AddCredit(ctx context.Context, userID string, amount int64, source, idempotencyKey string) (*CreditTransaction, error) {
    tx, err := l.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // 1. Verifica idempotência (evita duplicidade)
    var existingID string
    err = tx.QueryRowContext(ctx, `SELECT id FROM credits WHERE idempotency_key = $1`, idempotencyKey).Scan(&existingID)
    if err == nil {
        // Já processado, retorna a transação existente
        return l.getTransaction(ctx, existingID)
    } else if err != sql.ErrNoRows {
        return nil, err
    }

    // 2. Insere o crédito
    txn := &CreditTransaction{
        ID:        uuid.New().String(),
        UserID:    userID,
        Amount:    amount,
        Status:    "confirmed",
        Source:    source,
        CreatedAt: time.Now().Unix(),
    }
    _, err = tx.ExecContext(ctx, `
        INSERT INTO credits (id, user_id, amount, status, source, created_at, idempotency_key)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        txn.ID, txn.UserID, txn.Amount, txn.Status, txn.Source, txn.CreatedAt, idempotencyKey)
    if err != nil {
        return nil, err
    }

    // 3. Atualiza ou insere saldo do usuário
    _, err = tx.ExecContext(ctx, `
        INSERT INTO user_balance (user_id, balance, updated_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE SET balance = user_balance.balance + $2, updated_at = $3`,
        userID, amount, time.Now().Unix())
    if err != nil {
        return nil, err
    }

    // 4. Confirma transação
    if err := tx.Commit(); err != nil {
        return nil, err
    }

    // 5. Publica evento no Kafka (fora da transação, mas após o commit)
    l.publishEvent(txn)

    return txn, nil
}

func (l *Ledger) getTransaction(ctx context.Context, id string) (*CreditTransaction, error) {
    var txn CreditTransaction
    err := l.db.QueryRowContext(ctx, `SELECT id, user_id, amount, status, source, created_at FROM credits WHERE id = $1`, id).
        Scan(&txn.ID, &txn.UserID, &txn.Amount, &txn.Status, &txn.Source, &txn.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &txn, nil
}

func (l *Ledger) publishEvent(txn *CreditTransaction) {
    event := map[string]interface{}{
        "event":     "CreditoConfirmado",
        "user_id":   txn.UserID,
        "amount":    txn.Amount,
        "source":    txn.Source,
        "timestamp": txn.CreatedAt,
    }
    value, _ := json.Marshal(event)
    l.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
        Topic: "credits",
        Key:   []byte(txn.UserID),
        Value: value,
    })
}
