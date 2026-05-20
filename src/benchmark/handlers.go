package benchmark

import (
"encoding/json"
"net/http"
)

type Message struct {
Message string `json:"message"`
}

func JSONHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(Message{Message: "Hello, World!"})
}

func PlaintextHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "text/plain")
w.Write([]byte("Hello, World!"))
}

func FortuneHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "text/html")
w.Write([]byte(`<!DOCTYPE html><html><body><h1>Fortune</h1></body></html>`))
}

func DBHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"id": 1, "randomNumber": 42})
}

func QueriesHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode([]map[string]int{{"id": 1, "randomNumber": 42}, {"id": 2, "randomNumber": 43}})
}

func CachedQueryHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"id": 1, "randomNumber": 42})
}
