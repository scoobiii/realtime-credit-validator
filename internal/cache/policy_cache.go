package cache

import (
"sync"
"time"

"github.com/scoobiii/realtime-credit-validator/internal/models"
)

type PolicyCache struct {
mu    sync.RWMutex
items map[string]cacheItem
ttl   time.Duration
}

type cacheItem struct {
value     []models.Policy
expiresAt time.Time
}

func NewPolicyCache(ttl time.Duration) *PolicyCache {
return &PolicyCache{
make(map[string]cacheItem),
  ttl,
}
}

func (c *PolicyCache) Get(regulatorID string) ([]models.Policy, bool) {
c.mu.RLock()
item, found := c.items[regulatorID]
c.mu.RUnlock()
if found && time.Now().Before(item.expiresAt) {
 item.value, true
}
return nil, false
}

func (c *PolicyCache) Set(regulatorID string, policies []models.Policy) {
c.mu.Lock()
defer c.mu.Unlock()
c.items[regulatorID] = cacheItem{
    policies,
time.Now().Add(c.ttl),
}
}

func (c *PolicyCache) Invalidate(regulatorID string) {
c.mu.Lock()
defer c.mu.Unlock()
delete(c.items, regulatorID)
}
