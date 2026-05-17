package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
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

func (h *GatewayHandler) AddCredit(c *gin.Context) {
    authUserID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    var req CreditRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if req.UserID != authUserID {
        c.JSON(http.StatusForbidden, gin.H{"error": "user_id mismatch"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    txn, err := h.ledger.AddCredit(ctx, req.UserID, req.Amount, req.PaymentMethod, req.IdempotencyKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add credit"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":      "confirmed",
        "transaction": txn.ID,
        "message":     "Crédito disponível imediatamente. Sessão será renovada.",
    })
}
