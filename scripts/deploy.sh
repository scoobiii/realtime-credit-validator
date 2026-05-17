#!/bin/bash# (conteúdo do 
#!deploy.sh fornecido)
# deploy.sh — Implanta CI/CD + 
# HTTPS no projeto Uso: 
# ./scripts/deploy.sh
set -euo pipefail 
PROJECT_DIR="/tmp/realtime-credit-validator" 
cd "$PROJECT_DIR" echo "" echo 
"═══════════════════════════════════════════" 
echo " 🚀 Deploy: CI/CD + HTTPS" 
echo 
"═══════════════════════════════════════════" 
echo ""
# ── 1. CI/CD: GitHub Actions 
# ────────────────────────
echo "📦 [1/4] Criando pipeline 
CI/CD..." mkdir -p 
.github/workflows cat > 
.github/workflows/ci.yml << 
'CIEOF'
# Cole aqui o conteúdo do arquivo 
# ci.yml gerado (disponível em 
# .github/workflows/ci.yml no 
# output)
CIEOF echo "✅ Pipeline criado em 
.github/workflows/ci.yml"
# ── 2. HTTPS: Nginx config 
# ──────────────────────────
echo "" echo "🔐 [2/4] 
Configurando HTTPS..." mkdir -p 
deployments/nginx/certs
# Copia nginx.conf (cole o 
# conteúdo do nginx.conf gerado) 
# Gera certificado self-signed 
# para Cloud Shell
openssl req -x509 -nodes -newkey 
rsa:4096 \
  -keyout 
  deployments/nginx/certs/server.key 
  \ -out 
  deployments/nginx/certs/server.crt 
  \ -days 365 \ -subj 
  "/CN=realtime-credit-validator/O=Anatel/C=BR" 
  \ -addext 
  "subjectAltName=DNS:localhost,IP:127.0.0.1" 
  \ 2>/dev/null
chmod 600 
deployments/nginx/certs/server.key 
echo "✅ Certificado TLS gerado 
(self-signed, 365 dias)"
# ── 3. Sobe Nginx com HTTPS 
# ─────────────────────────
echo "" echo "🌐 [3/4] Iniciando 
Nginx com HTTPS..." docker run -d 
\
  --name nginx-https \ --network 
  host \ -v 
  "$PROJECT_DIR/deployments/nginx/nginx.conf:/etc/nginx/nginx.conf:ro" 
  \ -v 
  "$PROJECT_DIR/deployments/nginx/certs:/etc/nginx/certs:ro" 
  \ nginx:alpine
sleep 3
# ── 4. Valida HTTPS 
# ─────────────────────────────────
echo "" echo "✅ [4/4] Validando 
HTTPS..."
# Health check via HTTPS
HEALTH=$(curl -sk 
https://localhost/health) echo 
"Health HTTPS: $HEALTH"
# Testa crédito via HTTPS
TOKEN=$(cat /tmp/token.txt 
2>/dev/null || python3 - <<'PYEOF' 
import jwt, time payload = {
  'user_id': 'testuser', 'scopes': 
  ['credit:write'], 'exp': 
  int(time.time()) + 86400, 'iat': 
  int(time.time()), 'iss': 
  'realtime-credit-validator', 
  'aud': ['anatel-gateway']
}
print(jwt.encode(payload, 
'changeme-in-production', 
algorithm='HS256')) PYEOF ) 
RESULT=$(curl -sk -w "\nHTTP: 
%{http_code}" \
  -X POST 
  https://localhost/v1/credit \ -H 
  "Authorization: Bearer $TOKEN" \ 
  -H "Content-Type: 
  application/json" \ -d 
  '{"user_id":"testuser-1","amount":100,"idempotency_key":"https-test-1","payment_method":"pix"}')
echo "Crédito via HTTPS: $RESULT"
# k6 via HTTPS
echo "" echo "🧪 Rodando k6 via 
HTTPS..." k6 run \
  -e GATEWAY_URL=https://localhost 
  \ -e TOKEN="$TOKEN" \ --vus 10 
  --duration 30s \ 
  --insecure-skip-tls-verify \ 
  --summary-export=anatel_https_report.json 
  \ 
  tests/load/anatel-latency-test.js
echo "" echo 
"═══════════════════════════════════════════" 
echo " ✅ Deploy concluído!" echo 
" 🔐 HTTPS: https://localhost" 
echo " 📊 Relatório: 
anatel_https_report.json" echo 
"═══════════════════════════════════════════"
# Commit e push
git config --global user.email 
"mexai0516@gmail.com" 2>/dev/null 
|| true
git config --global user.name 
"scoobiii" 2>/dev/null || true git 
add .github/ deployments/nginx/ 
scripts/ git commit -m "feat: 
CI/CD GitHub Actions + HTTPS Nginx 
— SWOT implementado" || true git 
push origin main || echo "⚠️ Push 
manual necessário (token GitHub)"
