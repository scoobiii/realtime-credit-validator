#!/bin/bash

echo "📁 Criando estrutura de diretórios do projeto realtime-credit-validator..."

# Diretórios principais
mkdir -p .github/workflows
mkdir -p docs/api
mkdir -p src/wallet src/policy src/gateway/auth src/gateway/ratelimit src/gateway/handlers src/shared/messaging src/shared/db
mkdir -p deployments/k8s deployments/terraform
mkdir -p tests/integration tests/load
mkdir -p scripts

echo "📄 Criando arquivos de configuração e documentação..."

# .gitignore
cat > .gitignore << 'GITIGNORE'
# Binários e artefatos de build
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/
tmp/

# Arquivos de workspace do Go
go.work
go.work.sum

# Testes e cobertura
*.test
*.out
coverage.html
coverage.out

# Arquivos de ambiente com segredos (CRÍTICO)
.env
.env.local
.env.*.local

# Dependências
vendor/

# Arquivos de IDE
.vscode/
.idea/
*.swp
*.swo
*~

# Arquivos de sistema
.DS_Store
Thumbs.db

# Arquivos de banco de dados locais
*.db
*.db-shm
*.db-wal
GITIGNORE

# README.md
cat > README.md << 'README'
# realtime-credit-validator

Sistema de crédito em tempo real validado pela Anatel.
README
echo "# API Specification" > docs/api/openapi.yaml

# Makefile
cat > Makefile << 'MAKEFILE'
.PHONY: help build test clean

help:
	@echo "Comandos disponíveis: make build, make test, make clean"

build:
	@echo "Construindo o projeto..."

test:
	@echo "Executando testes..."

clean:
	@echo "Limpando arquivos temporários..."
MAKEFILE

echo "MIT License" > LICENSE
echo "# Workflow principal" > .github/workflows/ci.yml
echo "# Workflow de compliance Anatel" > .github/workflows/anatel-compliance.yml
echo "# Workflow de segurança" > .github/workflows/security-scan.yml
echo "# Documentação de compliance" > docs/anatel-compliance.md
echo "# Documentação de arquitetura" > docs/arquitetura-realtime.md

echo "⚙️ Criando arquivos de código (placeholders)..."

# Wallet
echo "package wallet\n\nfunc AddCredit() {}" > src/wallet/ledger.go
echo "package wallet\n\nfunc PublishEvent() {}" > src/wallet/events.go
echo "// Main do serviço wallet" > src/wallet/main.go
cat > src/wallet/Dockerfile << 'DOCKERFILE'
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o wallet ./src/wallet
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/wallet .
CMD ["./wallet"]
DOCKERFILE

# Policy
echo "package policy\n\nfunc UpdateEntitlement() {}" > src/policy/entitlement.go
echo "package policy\n\nfunc KillSession() {}" > src/policy/session.go
echo "// Main do serviço policy" > src/policy/main.go
echo "FROM golang:1.21-alpine AS builder\n...\nCMD [\"./policy\"]" > src/policy/Dockerfile

# Gateway
echo "// Main do serviço gateway" > src/gateway/main.go
echo "package auth\n\nfunc ValidateJWT() {}" > src/gateway/auth/jwt.go
echo "package ratelimit\n\nfunc AllowRequest() {}" > src/gateway/ratelimit/limiter.go
echo "package handlers\n\nfunc CreditHandler() {}" > src/gateway/handlers/credit.go
echo "FROM golang:1.21-alpine AS builder\n...\nCMD [\"./gateway\"]" > src/gateway/Dockerfile

# Shared
echo "package messaging\n\nfunc ConnectKafka() {}" > src/shared/messaging/kafka.go
echo "package db\n\nfunc ConnectPostgres() {}" > src/shared/db/postgres.go

echo "🐳 Criando arquivos de deployment..."

# docker-compose.yaml
cat > deployments/docker-compose.yaml << 'DOCKERCOMPOSE'
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: wallet
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  kafka:
    image: confluentinc/cp-kafka:latest
    ports:
      - "9092:9092"
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

volumes:
  postgres_data:
  redis_data:
DOCKERCOMPOSE

echo "# Namespace Kubernetes" > deployments/k8s/namespace.yaml
echo "# Deployment principal" > deployments/k8s/deployment.yaml
echo "# Service principal" > deployments/k8s/service.yaml
echo "# Ingress principal" > deployments/k8s/ingress.yaml
# Arquivos de segurança (placeholders)
echo "# Configuração Kafka mTLS" > deployments/k8s/kafka-mtls.yaml
echo "# Configuração PostgreSQL mTLS" > deployments/k8s/postgres-mtls.yaml
echo "# Configuração Redis mTLS" > deployments/k8s/redis-mtls.yaml
echo "# Configuração de criptografia K8s" > deployments/k8s/encryption-config.yaml
echo "# Configuração External Secrets" > deployments/k8s/external-secrets.yaml

echo "terraform {}" > deployments/terraform/main.tf

echo "🧪 Criando testes e scripts..."

# Testes
echo "package integration\n\nfunc TestRealtime() {}" > tests/integration/realtime_test.go
echo "# Script de teste de carga com Locust" > tests/load/locustfile.py
echo "# Script de teste de latência com k6" > tests/load/anatel-latency-test.js

# Scripts utilitários
echo "#!/bin/bash\necho 'Validando Anatel...'" > scripts/validate-anatel.sh
chmod +x scripts/validate-anatel.sh

echo "#!/bin/bash\necho 'Gerando certificados mTLS...'" > scripts/gen-mtls-certs.sh
chmod +x scripts/gen-mtls-certs.sh

echo "#!/bin/bash\necho 'Executando segurança...'" > scripts/security-scan.sh
chmod +x scripts/security-scan.sh

echo "-- Comandos SQL iniciais" > scripts/seed-data.sql

echo "✅ Estrutura do projeto criada com sucesso!"
echo "📂 Próximos passos:"
echo "1. Revise os arquivos gerados, especialmente os placeholders em src/"
echo "2. Execute 'go mod init github.com/scoobiii/realtime-credit-validator' para iniciar o módulo Go"
echo "3. Faça o commit inicial: git add . && git commit -m 'Sprint 0: Estrutura inicial do projeto'"
echo "4. Envie para o GitHub: git push origin main"
