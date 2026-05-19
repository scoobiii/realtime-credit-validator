package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lib/pq"
    "github.com/scoobiii/realtime-credit-validator/src/wallet"
)

type CreditRequest struct {
    UserID         string `json:"user_id" binding:"required"`
    Amount         int64  `json:"amount" binding:"required,min=1"`
    IdempotencyKey string `json:"idempotency_key" binding:"required"`
    PaymentMethod  string `json:"payment_method" binding:"oneof=pix credit_card"`
}

type GatewayHandler struct {
    ledger *wallet.Ledger
}

func NewGatewayHandler(ledger *wallet.Ledger) *GatewayHandler {
    return &GatewayHandler{ledger: ledger}
}

// isTransientError verifica se o erro merece retry
func isTransientError(err error) bool {
    if err == nil {
        return false
    }
    // PostgreSQL serialization_failure, deadlock, out of memory
    if pqErr, ok := err.(*pq.Error); ok {
        switch pqErr.Code {
        case "40001", "40P01", "53200":
            return true
        }
    }
    // Erros de rede / timeout (Redis, Kafka)
    if err.Error() == "redis: nil" ||
       err.Error() == "context deadline exceeded" ||
       err.Error() == "connection refused" ||
       err.Error() == "i/o timeout" {
        return true
    }
    return false
}

func (h *GatewayHandler) AddCredit(c *gin.Context) {
    _, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
        return
    }

    var req CreditRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    const maxRetries = 3
    var txn *wallet.CreditTransaction
    var err error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        txn, err = h.ledger.AddCredit(ctx, req.UserID, req.Amount, req.PaymentMethod, req.IdempotencyKey)
        cancel()

        if err == nil {
            c.JSON(http.StatusOK, gin.H{
                "status":      "confirmed",
                "transaction": txn.ID,
                "message":     "Crédito disponível imediatamente. Sessão será renovada.",
            })
            return
        }

        if isTransientError(err) && attempt < maxRetries {
            backoff := time.Duration(10*(1<<uint(attempt-1))) * time.Millisecond // 10,20,40ms
            time.Sleep(backoff)
            continue
        }
        break
    }

    if isTransientError(err) {
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "serviço temporariamente indisponível. Tente novamente."})
    } else {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add credit: " + err.Error()})
    }
}
