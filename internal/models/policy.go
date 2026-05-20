package models

import (
"encoding/json"
"time"

"github.com/google/uuid"
)

type Policy struct {
ID          uuid.UUID       `json:"id"`
RegulatorID string          `json:"regulator_id"`
MarketID    string          `json:"market_id"`
RulesJSON   json.RawMessage `json:"rules_json"`
Version     int             `json:"version"`
IsActive    bool            `json:"is_active"`
CreatedAt   time.Time       `json:"created_at"`
UpdatedAt   time.Time       `json:"updated_at"`
}

type CreatePolicyRequest struct {
RegulatorID string          `json:"regulator_id" binding:"required"`
MarketID    string          `json:"market_id" binding:"required"`
RulesJSON   json.RawMessage `json:"rules_json" binding:"required"`
}

type UpdatePolicyRequest struct {
RulesJSON json.RawMessage `json:"rules_json" binding:"required"`
IsActive  *bool           `json:"is_active"`
}
