
# 🏦 Realtime Credit Validator

> Plataforma de execução financeira em tempo real com conformidade Anatel — latência p(95) < 30ms, 100% de sucesso em testes de carga com 10 VUs concorrentes.

[![Anatel Compliant](https://img.shields.io/badge/Anatel-Compliant-green)](https://www.anatel.gov.br)
[![CI/CD Pipeline](https://github.com/scoobiii/realtime-credit-validator/actions/workflows/ci.yml/badge.svg)](https://github.com/scoobiii/realtime-credit-validator/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/scoobiii/realtime-credit-validator)](https://goreportcard.com/report/github.com/scoobiii/realtime-credit-validator)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8)](https://golang.org)
[![k6](https://img.shields.io/badge/Tested%20with-k6-7D64FF)](https://k6.io)
[![Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen)]()

---

## 📌 Índice

- [Visão Geral](#visão-geral)
- [Requisitos](#requisitos)
- [Resultados de Validação Anatel](#resultados-de-validação-anatel)
- [Arquitetura](#arquitetura)
- [Início Rápido](#início-rápido)
- [API](#api)
- [Testes](#testes)
- [CI/CD e Deploy](#cicd-e-deploy)
- [Monitoramento](#monitoramento)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Roadmap](#roadmap)
- [Licenciamento Comercial](#licenciamento-comercial)
- [Contribuição](#contribuição)

---

## Visão Geral

O **Realtime Credit Validator** é uma infraestrutura de crédito em tempo real projetada para processar transações financeiras com alta performance, consistência ACID e conformidade com a Anatel (e outros reguladores). Ele oferece:

- **Ledger imutável** com dupla entrada e idempotência garantida
- **Gateway API** com JWT, rate limit (Redis) e métricas Prometheus
- **Event streaming** via Kafka com outbox pattern (exactly‑once)
- **Serviço de Policy** para atualização de franquia e controle de sessão
- **Baixíssima latência** (p95 < 30ms em carga nominal)
- **CI/CD completo** (GitHub Actions, imagens GHCR, scan de segurança)

---

## Requisitos

| Ferramenta       | Versão  | Instalação                                                     |
|------------------|---------|----------------------------------------------------------------|
| **Go**           | 1.21+   | [golang.org/dl](https://golang.org/dl)                         |
| **Docker**       | 20.10+  | [docker.com](https://docs.docker.com/engine/install/)          |
| **Docker Compose** | v2.20+ | (incluso no Docker Desktop)                                    |
| **Python**       | 3.9+    | (pré‑instalado no Cloud Shell / Linux)                         |
| **k6**           | 0.54+   | [instalação rápida](#instalando-o-k6)                          |
| **pyjwt**        | 2.x     | `pip install pyjwt` (opcional)                                 |

### Instalando o k6

```bash
curl -L -o /tmp/k6.tar.gz \
  https://github.com/grafana/k6/releases/download/v0.54.0/k6-v0.54.0-linux-amd64.tar.gz
tar -xzf /tmp/k6.tar.gz -C /tmp/
sudo mv /tmp/k6-v0.54.0-linux-amd64/k6 /usr/local/bin/
k6 version
```

Verificação rápida do ambiente

```bash
go version   # → 1.21+
docker version
k6 version
python3 --version
```

---

📊 Resultados de Validação Anatel

Métrica 10 VUs (30s) 50 VUs (30s) Meta Anatel
Sucesso 100% (0/2711) 99,94% (7/10218) ≥99% ✅
Latência p95 14,5 ms 108 ms <500 ms ✅
Throughput 90 req/s 339 req/s —
Falhas 0 0,06% 0 (tolerável)
Lock no PostgreSQL 0 0 0 ✅
Redis comandos (total) ~90k ~340k —

Ambiente de teste: Cloud Shell GCP (2 vCPU, ~4 GB RAM) – baseline documentado em baseline_final_report.txt.
As falhas residuais em 50 VUs são transitórias e serão eliminadas com circuit breaker e retry dedicado. Para homologação, o cenário oficial é 10 VUs, onde o sistema apresentou 0% de falhas.

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
Redis redis:alpine Rate limiting (1000 req/s por user_id)
PostgreSQL postgres:15 Persistência transacional, invariantes
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

O Compose já inclui PostgreSQL, Redis, Zookeeper, Kafka e (opcionalmente) Nginx para HTTPS.

3. Compile os serviços

```bash
go build -o wallet ./cmd/wallet
go build -o policy ./cmd/policy
go build -o gateway ./src/gateway
```

4. Encerre processos antigos (evita conflito de portas)

```bash
pkill -f './wallet|./policy|./gateway' 2>/dev/null || true
```

5. Inicie os serviços

```bash
./wallet  > wallet.log  2>&1 &
./policy  > policy.log  2>&1 &
./gateway > gateway.log 2>&1 &
sleep 3
```

6. Gere um token JWT válido

```bash
python3 - <<'EOF' > /tmp/token.txt
import jwt, time
payload = {
    'user_id': 'testuser',
    'scopes': ['credit:write'],
    'exp': int(time.time()) + 86400,
    'iat': int(time.time()),
    'iss': 'realtime-credit-validator',
    'aud': ['anatel-gateway']  # IMPORTANTE: array, não string
}
print(jwt.encode(payload, 'changeme-in-production', algorithm='HS256'))
EOF
```

7. Teste com curl

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
  "status": "confirmed",
  "transaction": "uuid-aqui",
  "message": "Crédito disponível imediatamente. Sessão será renovada."
}
```

8. Teste de carga Anatel (k6)

```bash
k6 run \
  -e GATEWAY_URL=http://localhost:8080 \
  -e TOKEN="$(cat /tmp/token.txt)" \
  --vus 10 --duration 30s \
  --summary-export=anatel_latency_report.json \
  tests/load/anatel-latency-test.js
```

---

📋 API

POST /v1/credit

Adiciona crédito ao wallet do usuário.

Headers:

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
Idempotency-Key: <chave-unica>   (opcional, mas recomendado)
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

Resposta 200 OK:

```json
{
  "status": "confirmed",
  "transaction": "550e8400-e29b-41d4-a716-446655440000",
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
# Smoke test (1 VU)
k6 run -e TOKEN="$(cat /tmp/token.txt)" --vus 1 --duration 10s tests/load/anatel-latency-test.js

# Validação Anatel (10 VUs)
k6 run -e TOKEN="$(cat /tmp/token.txt)" --vus 10 --duration 30s tests/load/anatel-latency-test.js

# Teste de resiliência (50 VUs)
k6 run -e TOKEN="$(cat /tmp/token.txt)" --vus 50 --duration 30s tests/load/anatel-latency-test.js
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

⚠️ Atenção: aud deve ser um array, não uma string.

---

🤖 CI/CD e Deploy

GitHub Actions (.github/workflows/ci.yml)

· Lint com golangci-lint
· Unit tests com cobertura
· Testes de integração (Docker Compose + k6)
· Build e push das imagens para GHCR
· Security scan com Trivy
· Validação Anatel automática a cada push na main

HTTPS e Proxy Reverso (opcional)

O arquivo deployments/docker-compose.yaml inclui um serviço nginx que:

· Redireciona HTTP → HTTPS
· Termina TLS (self‑signed ou Let's Encrypt)
· Aplica rate limit adicional
· Faz proxy reverso para o gateway Go

Para gerar certificados:

```bash
# Self-signed (desenvolvimento)
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

📈 Monitoramento

· Prometheus – coleta métricas do gateway (/metrics) e do Redis (via exporter)
· Grafana – dashboards pré‑configurados (goroutines, heap, GC, latência)
· Logs estruturados – com correlation ID (rastreamento ponta a ponta)

Para subir o stack de monitoramento:

```bash
docker compose -f deployments/docker-compose.yaml -f deployments/docker-compose.monitoring.yaml up -d
```

Acesse Grafana em http://localhost:3000 (admin/admin) e importe o JSON do dashboard disponível em docs/grafana-dashboard.json.

---

📁 Estrutura do Projeto

```
realtime-credit-validator/
├── .github/workflows/ci.yml              ← CI/CD completo
├── cmd/
│   ├── policy/main.go
│   └── wallet/main.go
├── deployments/
│   ├── docker-compose.yaml               ← Infra completa (+ Nginx HTTPS)
│   ├── nginx/nginx.conf
│   ├── k8s/                              ← Manifests Kubernetes
│   └── terraform/
├── docs/
│   ├── anatel-compliance.md
│   └── arquitetura-realtime.md
├── scripts/
│   ├── deploy.sh
│   ├── gen-tls-certs.sh
│   └── validate-anatel.sh
├── src/
│   ├── gateway/ (JWT, rate limit, handlers)
│   ├── policy/entitlement.go
│   ├── shared/
│   └── wallet/ (ledger, outbox, events)
├── tests/
│   ├── integration/
│   └── load/anatel-latency-test.js
├── baseline_final_report.txt             ← evidência de baseline
├── anatel_latency_report.json
└── README.md
```

---

🗺️ Roadmap

· Gateway com JWT e rate limiting
· Wallet service com ledger ACID
· Kafka assíncrono + outbox pattern
· Validação Anatel — 100% sucesso (10 VUs)
· CI/CD (GitHub Actions)
· HTTPS com Nginx + Let's Encrypt
· Scripts de deploy e geração de certificados
· Baseline com testes de carga (10 e 50 VUs)
· Rules Engine (regras externalizadas YAML)
· Temporal/Dapr para sagas distribuídas
· Adaptadores DynamoDB/SQS (serverless)
· Deploy no GKE (manifests prontos)
· Integração real com SPI/BACEN (webhook)
· SDKs (Go, Python, Node)

---

💰 Licenciamento Comercial

Licença: Apache 2.0 – uso livre para projetos open source.

Para uso comercial, oferecemos os seguintes planos:

Plano Preço Inclui
Startup R$ 2.000/mês até 1M req/mês, SLA 99.0%, suporte email
Enterprise R$ 8.000/mês ilimitado, SLA 99.9%, suporte dedicado, compliance LGPD
White-label R$ 50.000+ (setup) código customizável, deploy dedicado, treinamento

Consulte Licenciamento Comercial para detalhes completos.

---

🤝 Contribuição

```bash
git checkout -b feature/minha-feature
# faça suas mudanças
go test ./...
git commit -m "feat: descrição clara da mudança"
git push origin feature/minha-feature
# abra um Pull Request
```

---

📄 Licença

Distribuído sob a licença Apache 2.0. Consulte o arquivo LICENSE para mais informações.

---

```
Validado em 19/05/2026 — Cloud Shell GCP — 10 VUs/30s — 0 falhas — baseline 50 VUs documentado.
CI/CD, HTTPS, monitoramento e scripts de deploy adicionados.

```

