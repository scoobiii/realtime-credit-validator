

# 🏦 Realtime Credit Validator

> Validador de crédito em tempo real com conformidade Anatel — latência p(95) < 30ms, 100% de sucesso em testes de carga com 10 VUs concorrentes.

[![Anatel Compliant](https://img.shields.io/badge/Anatel-Compliant-green)](https://www.anatel.gov.br)
[![CI/CD Pipeline](https://github.com/scoobiii/realtime-credit-validator/actions/workflows/ci.yml/badge.svg)](https://github.com/scoobiii/realtime-credit-validator/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8)](https://golang.org)
[![k6](https://img.shields.io/badge/Tested%20with-k6-7D64FF)](https://k6.io)


## 📋 Requisitos

| Ferramenta | Versão | Instalação |
|------------|--------|-------------|
| **Go** | 1.21+ | [golang.org/dl](https://golang.org/dl) |
| **Docker** | 20.10+ | [docker.com](https://docs.docker.com/engine/install/) |
| **Docker Compose** | v2.20+ | (já incluso no Docker Desktop) |
| **Python** | 3.9+ | (pré‑instalado no Cloud Shell / Linux) |
| **k6** | 0.54+ | [instalação rápida](#instalando-o-k6) |
| **pyjwt** | 2.x | `pip install pyjwt` (opcional, apenas para geração manual de token) |

### Instalando o k6 (caso não tenha)

```bash
# Linux (via .deb)
sudo gpg -k && sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt update && sudo apt install k6

# Ou baixar o binário diretamente
curl -L -o k6.tar.gz https://github.com/grafana/k6/releases/download/v0.54.0/k6-linux-amd64.tar.gz
tar -xzf k6.tar.gz
sudo mv k6-v0.54.0-linux-amd64/k6 /usr/local/bin/
```

Verificação rápida

```bash
go version   # → 1.21+
docker version
k6 version
python3 --version
```

---

📊 Resultados de Validação Anatel

Métrica Resultado Meta
Taxa de sucesso 100% 99% ✅
Latência média 13.47ms < 500ms ✅
Latência p(90) 22.13ms < 500ms ✅
Latência p(95) 27.44ms < 500ms ✅
Falhas 0 / 2620 0 ✅
VUs concorrentes 10 10 ✅
Throughput 87 req/s —
Duração do teste 30s 30s ✅

---

🏗️ Arquitetura

```
┌─────────────┐     JWT/TLS      ┌─────────────────┐
│   Cliente   │ ──────────────▶  │   API Gateway   │
│  (k6/curl)  │                  │   :8080 (Gin)   │
└─────────────┘                  └────────┬────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
             ┌──────▼──────┐    ┌─────────▼──────┐   ┌─────────▼──────┐
             │    Redis     │    │   PostgreSQL   │   │     Kafka      │
             │  Rate Limit  │    │  Wallet/Ledger │   │  Event Stream  │
             │  1000 req/s  │    │  pool: 50 conn │   │  Async/Topic   │
             └─────────────┘    └────────────────┘   └────────────────┘
                                          │
                              ┌───────────▼───────────┐
                              │    Wallet Service      │
                              │    Policy Service      │
                              └───────────────────────┘
```

Componentes

Serviço Tecnologia Função
Gateway Go + Gin Auth JWT, Rate Limit, Roteamento
Wallet Go Ledger, Saldo, Crédito
Policy Go Regras de negócio, Entitlements
Redis redis:alpine Rate limiting (1000 req/s por usuário)
PostgreSQL postgres:15 Persistência transacional
Kafka confluentinc/cp-kafka Event sourcing assíncrono
Zookeeper confluentinc/cp-zookeeper Coordenação Kafka
Nginx (opcional) nginx:alpine Proxy reverso com HTTPS

---

🚀 Início Rápido

1. Clone o repositório

```bash
git clone https://github.com/scoobiii/realtime-credit-validator.git
cd realtime-credit-validator
```

2. Suba a infraestrutura com Docker Compose

```bash
docker compose -f deployments/docker-compose.yaml up -d
sleep 10
docker ps
```

Nota: O Compose já inclui PostgreSQL, Redis, Zookeeper, Kafka e (opcionalmente) Nginx para HTTPS.

3. Compile os serviços

```bash
go build -o wallet ./cmd/wallet
go build -o policy ./cmd/policy
go build -o gateway ./src/gateway
```

4. Inicie os serviços

```bash
./wallet  > wallet.log  2>&1 &
./policy  > policy.log  2>&1 &
./gateway > gateway.log 2>&1 &
sleep 3
```

5. Gere um token JWT válido

```bash
python3 - <<'EOF' > /tmp/token.txt
import jwt, time
payload = {
  'user_id': 'testuser',
  'scopes': ['credit:write'],
  'exp': int(time.time()) + 86400,
  'iat': int(time.time()),
  'iss': 'realtime-credit-validator',
  'aud': ['anatel-gateway']          # IMPORTANTE: array, não string
}
print(jwt.encode(payload, 'changeme-in-production', algorithm='HS256'))
EOF
```

6. Teste com curl

```bash
curl -s -w "\nHTTP: %{http_code}\n" \
  -X POST http://localhost:8080/v1/credit \
  -H "Authorization: Bearer $(cat /tmp/token.txt)" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"testuser-1","amount":100,"idempotency_key":"test-1","payment_method":"pix"}'
```

Resposta esperada:

```json
{
  "message": "Crédito disponível imediatamente. Sessão será renovada.",
  "status": "confirmed",
  "transaction": "uuid-aqui"
}
```

7. Teste de carga Anatel (k6)

```bash
k6 run \
  -e GATEWAY_URL=http://localhost:8080 \
  -e TOKEN="$(cat /tmp/token.txt)" \
  --vus 10 --duration 30s \
  --summary-export=anatel_latency_report.json \
  tests/load/anatel-latency-test.js
```

---

⚙️ Variáveis de Ambiente

Variável Padrão Descrição
JWT_SECRET changeme-in-production Chave de assinatura JWT
DB_URL postgres://user:password@postgres:5432/wallet?sslmode=disable URL do PostgreSQL
REDIS_ADDR redis:6379 Endereço do Redis
KAFKA_BROKERS kafka:9092 Brokers Kafka
PORT 8080 Porta do gateway

---

🔐 Autenticação JWT

O token JWT deve conter as seguintes claims:

```json
{
  "user_id": "string",
  "scopes": ["credit:write"],
  "iss": "realtime-credit-validator",
  "aud": ["anatel-gateway"],
  "exp": 1234567890
}
```

⚠️ Importante: aud deve ser um array, não uma string.

---

📋 API

POST /v1/credit

Adiciona crédito ao wallet do usuário.

Headers:

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

Body:

```json
{
  "user_id": "testuser-1",
  "amount": 100,
  "idempotency_key": "chave-unica-por-operacao",
  "payment_method": "pix"
}
```

Resposta 200:

```json
{
  "status": "confirmed",
  "transaction": "uuid",
  "message": "Crédito disponível imediatamente. Sessão será renovada."
}
```

Códigos de resposta:

Código Significado
200 Crédito confirmado
401 Token inválido ou ausente
403 Scope insuficiente (aud errado)
429 Rate limit excedido (1000 req/s)
500 Erro interno

GET /health

```bash
curl http://localhost:8080/health
# {"status": "ok"}
```

---

🧪 Testes

Unitários e integração

```bash
go test ./...
```

Teste de carga (k6)

```bash
# 1 VU – smoke test
k6 run -e TOKEN="$(cat /tmp/token.txt)" --vus 1 --duration 10s tests/load/anatel-latency-test.js

# 10 VUs – validação Anatel
k6 run -e TOKEN="$(cat /tmp/token.txt)" --vus 10 --duration 30s tests/load/anatel-latency-test.js
```

---

🔧 Decisões Técnicas

Rate Limiting no Redis

O rate limiter usa sliding window com Lua script no Redis, garantindo atomicidade mesmo com múltiplos pods. Limite: 1000 req/s por user_id.

Kafka Assíncrono

```go
Async: true
RequiredAcks: kafka.RequireOne
```

Eventos são publicados sem bloquear a resposta HTTP, reduzindo latência de ~50ms para ~5ms.

Pool de Conexões PostgreSQL

```go
db.SetMaxOpenConns(50)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

Evita contenção com 10+ VUs concorrentes.

JWT com aud como array

O gateway espera Audience: []string{"anatel-gateway"}. Gere sempre o token com aud como lista.

---

📁 Estrutura do Projeto

```
realtime-credit-validator/
├── .github/workflows/
│   └── ci.yml                ← CI/CD completo (testes, build, push, segurança)
├── cmd/
│   ├── policy/main.go
│   └── wallet/main.go
├── deployments/
│   ├── docker-compose.yaml   ← Infra completa (+ Nginx HTTPS)
│   ├── nginx/
│   │   ├── nginx.conf        ← Proxy reverso com TLS
│   │   └── certs/            ← Certificados (self-signed ou Let's Encrypt)
│   ├── k8s/                  → Manifests Kubernetes
│   └── terraform/            → Infra como código
├── docs/
│   ├── anatel-compliance.md
│   ├── api/openapi.yaml
│   └── arquitetura-realtime.md
├── scripts/
│   ├── deploy.sh             ← Automatiza CI + HTTPS
│   ├── gen-tls-certs.sh      ← Gera certificados TLS
│   ├── gen-token.sh
│   └── validate-anatel.sh
├── src/
│   ├── gateway/
│   │   ├── auth/jwt.go
│   │   ├── handlers/credit.go
│   │   ├── ratelimit/limiter.go
│   │   └── main.go
│   ├── policy/entitlement.go
│   ├── shared/
│   │   ├── db/postgres.go
│   │   └── messaging/kafka.go
│   └── wallet/
│       ├── events.go        ← Kafka async
│       ├── ledger.go
│       └── events.go
├── tests/
│   ├── integration/realtime_test.go
│   └── load/
│       └── anatel-latency-test.js
├── anatel_latency_report.json   ← evidência do teste
├── anatel_evidence_gateway.log  ← logs do gateway
├── anatel_containers.txt        ← prova da arquitetura
└── README.md
```

---

🤖 CI/CD e Deploy

GitHub Actions (.github/workflows/ci.yml)

· Lint com golangci-lint
· Unit tests com cobertura
· Testes de integração (Docker Compose + k6)
· Build e push das imagens para GHCR
· Security scan com Trivy
· Validação Anatel automática a cada push na main

HTTPS e Proxy Reverso

O arquivo deployments/docker-compose.yaml inclui um serviço nginx que:

  · Redireciona HTTP → HTTPS
  · Termina TLS (certificados self-signed ou Let's Encrypt)
  · Aplica rate limit adicional
  · Faz proxy reverso para o gateway Go

Para gerar certificados:

```bash
# Self-signed (para desenvolvimento)
./scripts/gen-tls-certs.sh self-signed

# Let's Encrypt (produção, exige domínio público)
./scripts/gen-tls-certs.sh letsencrypt api.meudominio.com br@meudominio.com
```

Deploy rápido

```bash
./scripts/deploy.sh
```

Este script configura CI, gera certificados, sobe o ambiente com HTTPS e roda o teste k6 via TLS.

---

📈 Roadmap

· Gateway com JWT e rate limiting
· Wallet service com ledger
· Kafka assíncrono
· Validação Anatel — 100% sucesso
· CI/CD (GitHub Actions)
· HTTPS com Nginx + Let's Encrypt
· Scripts de deploy e geração de certificados
· Deploy no GKE (manifests em /deployments/k8s)
· Dashboard Grafana com métricas k6
· Integração BACEN SPI (Pix real)
· SDK cliente (Go, Python, Node)


💰 Licença e Uso Comercial

Licença: Apache 2.0

Uso livre para projetos open source. Para uso comercial:
  
  Tier Preço Inclui
  Startup R$ 2.000/mês até 1M req/mês, suporte email
  Enterprise R$ 8.000/mês ilimitado, SLA 99.9%, suporte dedicado
  White-label R$ 50.000 único código fonte + customização + treinamento
  
Contato: abra uma issue ou env  ie email.

---

🤝 Contribuindo

```bash
git checkout -b feature/minha-feature
# faça suas mudanças
go test ./...
git commit -m "feat: descrição clara da mudança"
git push origin feature/minha-feature
# abra um Pull Request
```

---

Validado em 17/05/2026 — Cloud Shell GCP — 10 VUs — 30s — 0 falhas
CI/CD, HTTPS e scripts de deploy adicionados em 17/05/2026

```
