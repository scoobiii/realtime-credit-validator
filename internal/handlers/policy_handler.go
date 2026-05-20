package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
"github.com/scoobiii/realtime-credit-validator/internal/cache"
"github.com/scoobiii/realtime-credit-validator/internal/models"
"github.com/scoobiii/realtime-credit-validator/internal/repository"
)

type PolicyHandler struct {
repo  *repository.PolicyRepository
cache *cache.PolicyCache
}

func NewPolicyHandler(repo *repository.PolicyRepository, c *cache.PolicyCache) *PolicyHandler {
return &PolicyHandler{repo: repo, cache: c}
}

func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
var req models.CreatePolicyRequest
if err := c.ShouldBindJSON(&req); err != nil {
(http.StatusBadRequest, gin.H{"error": err.Error()})

}
policy := &models.Policy{
         uuid.New(),
req.RegulatorID,
   req.MarketID,
:   req.RulesJSON,
:     1,
   true,
}
if err := h.repo.Create(policy); err != nil {
(http.StatusInternalServerError, gin.H{"error": "failed to create policy"})

}
h.cache.Invalidate(policy.RegulatorID)
c.JSON(http.StatusCreated, policy)
}

func (h *PolicyHandler) GetPoliciesByRegulator(c *gin.Context) {
regulatorID := c.Query("regulator_id")
if regulatorID == "" {
(http.StatusBadRequest, gin.H{"error": "regulator_id required"})

}
if policies, ok := h.cache.Get(regulatorID); ok {
(http.StatusOK, policies)

}
policies, err := h.repo.GetActiveByRegulator(regulatorID)
if err != nil {
(http.StatusInternalServerError, gin.H{"error": "failed to fetch policies"})

}
h.cache.Set(regulatorID, policies)
c.JSON(http.StatusOK, policies)
}
