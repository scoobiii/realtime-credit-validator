package repository

import (
"database/sql"
"errors"

"github.com/google/uuid"
"github.com/scoobiii/realtime-credit-validator/internal/models"
)

type PolicyRepository struct {
db *sql.DB
}

func NewPolicyRepository(db *sql.DB) *PolicyRepository {
return &PolicyRepository{db: db}
}

func (r *PolicyRepository) Create(p *models.Policy) error {
query := `
SERT INTO policies (id, regulator_id, market_id, rules_json, version, is_active)
($1, $2, $3, $4, $5, $6)
ING created_at, updated_at
`
return r.db.QueryRow(query, p.ID, p.RegulatorID, p.MarketID, p.RulesJSON, p.Version, p.IsActive).
(&p.CreatedAt, &p.UpdatedAt)
}

func (r *PolicyRepository) GetByID(id uuid.UUID) (*models.Policy, error) {
query := `SELECT id, regulator_id, market_id, rules_json, version, is_active, created_at, updated_at
          FROM policies WHERE id = $1`
var p models.Policy
err := r.db.QueryRow(query, id).Scan(
&p.RegulatorID, &p.MarketID, &p.RulesJSON,
, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
)
if err == sql.ErrNoRows {
 nil, errors.New("policy not found")
}
return &p, err
}

func (r *PolicyRepository) GetActiveByRegulator(regulatorID string) ([]models.Policy, error) {
query := `SELECT id, regulator_id, market_id, rules_json, version, is_active, created_at, updated_at
          FROM policies WHERE regulator_id = $1 AND is_active = true ORDER BY market_id`
rows, err := r.db.Query(query, regulatorID)
if err != nil {
 nil, err
}
defer rows.Close()
var policies []models.Policy
for rows.Next() {
p models.Policy
err := rows.Scan(&p.ID, &p.RegulatorID, &p.MarketID, &p.RulesJSON,
, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
 nil, err
= append(policies, p)
}
return policies, nil
}

func (r *PolicyRepository) Update(id uuid.UUID, req *models.UpdatePolicyRequest) error {
query := `UPDATE policies SET rules_json = $1, is_active = $2, version = version + 1, updated_at = NOW() WHERE id = $3`
_, err := r.db.Exec(query, req.RulesJSON, req.IsActive, id)
return err
}

func (r *PolicyRepository) Delete(id uuid.UUID) error {
_, err := r.db.Exec(`DELETE FROM policies WHERE id = $1`, id)
return err
}
