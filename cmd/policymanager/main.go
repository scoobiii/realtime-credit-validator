package main

import (
"database/sql"
"log"
"os"
"time"

"github.com/gin-gonic/gin"
_ "github.com/lib/pq"
"github.com/scoobiii/realtime-credit-validator/internal/cache"
"github.com/scoobiii/realtime-credit-validator/internal/handlers"
"github.com/scoobiii/realtime-credit-validator/internal/repository"
)

func main() {
dbURL := os.Getenv("DB_URL")
if dbURL == "" {
= "postgres://user:password@localhost:5432/policymanager?sslmode=disable"
}
db, err := sql.Open("postgres", dbURL)
if err != nil {
db.Close()
db.SetMaxOpenConns(50)

policyRepo := repository.NewPolicyRepository(db)
policyCache := cache.NewPolicyCache(5 * time.Minute)
policyHandler := handlers.NewPolicyHandler(policyRepo, policyCache)

r := gin.Default()
r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

api := r.Group("/api/v1/policies")
{
policyHandler.CreatePolicy)
policyHandler.GetPoliciesByRegulator)
}

port := os.Getenv("PORT")
if port == "" {
= "8082"
}
log.Printf("Policy Manager listening on :%s", port)
r.Run(":" + port)
}
