#!/bin/bash
set -e
cd "$(dirname "$0")/.."

echo "🚀 Gerando certificados..."
./scripts/gen-tls-certs.sh

echo "🔐 Gerando token JWT..."
python3 -c "
import jwt, time
payload = {
    'user_id': 'testuser',
    'scopes': ['credit:write'],
    'exp': int(time.time()) + 86400,
    'iat': int(time.time()),
    'iss': 'realtime-credit-validator',
    'aud': ['anatel-gateway']
}
print(jwt.encode(payload, 'changeme-in-production', algorithm='HS256'))
" > /tmp/token.txt

if [ ! -s /tmp/token.txt ]; then
    echo "❌ Erro: token não gerado"
    exit 1
fi
echo "✅ Token gerado (tamanho: $(wc -c < /tmp/token.txt) bytes)"

echo "📦 Verificando k6..."
if ! command -v k6 &> /dev/null; then
    echo "Instalando k6 (via apt)..."
    sudo apt update && sudo apt install -y k6
fi

echo "🐳 Subindo infra + monitoring..."
docker compose -f deployments/docker-compose.yaml -f deployments/docker-compose.monitoring.yaml up -d

echo "⏳ Aguardando PostgreSQL ficar pronto..."
sleep 5

echo "📦 Executando migrations (criação das tabelas)..."
docker exec -i deployments-postgres-1 psql -U user -d wallet << 'SQL'
CREATE TABLE IF NOT EXISTS credits (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL,
    source VARCHAR(20) NOT NULL,
    created_at BIGINT NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS user_balance (
    user_id VARCHAR(100) PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_credits_user_id ON credits(user_id);
CREATE INDEX IF NOT EXISTS idx_credits_created_at ON credits(created_at);
CREATE INDEX IF NOT EXISTS idx_credits_idempotency ON credits(idempotency_key);
SQL

echo "🔨 Compilando serviços..."
go build -o wallet ./cmd/wallet
go build -o policy ./cmd/policy
go build -o gateway ./src/gateway

echo "🔄 Matando processos antigos..."
pkill -f './wallet|./policy|./gateway' 2>/dev/null || true

echo "🚀 Iniciando wallet..."
./wallet > wallet.log 2>&1 &
echo "🚀 Iniciando policy..."
./policy > policy.log 2>&1 &
echo "🚀 Iniciando gateway..."
./gateway > gateway.log 2>&1 &
sleep 3

echo "🏥 Health check..."
curl -f http://localhost:8080/health || exit 1

echo "🧪 Teste k6 (10 VUs)..."
k6 run -e GATEWAY_URL=http://localhost:8080 -e TOKEN="$(cat /tmp/token.txt)" --vus 10 --duration 10s tests/load/anatel-latency-test.js

echo "✅ Deploy concluído. Gateway em http://localhost:8080"
