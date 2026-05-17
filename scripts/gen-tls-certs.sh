#!/bin/bash# (conteúdo do 
#!gen-tls-certs.sh fornecido)
# gen-tls-certs.sh — Gera 
# certificados TLS para o gateway 
# Uso: ./scripts/gen-tls-certs.sh 
# [self-signed|letsencrypt]
set -euo pipefail 
MODE="${1:-self-signed}" 
CERT_DIR="deployments/nginx/certs" 
mkdir -p "$CERT_DIR" case "$MODE" 
in
  # ──────────────────────────────────────────────── 
  # Self-signed — para dev/Cloud 
  # Shell 
  # ────────────────────────────────────────────────
  self-signed) echo "🔐 Gerando 
    certificado self-signed..." 
    openssl req -x509 -nodes 
    -newkey rsa:4096 \
      -keyout 
      "$CERT_DIR/server.key" \ 
      -out "$CERT_DIR/server.crt" 
      \ -days 365 \ -subj 
      "/CN=realtime-credit-validator/O=Anatel/C=BR" 
      \ -addext 
      "subjectAltName=DNS:localhost,IP:127.0.0.1"
    echo "✅ Certificado gerado em 
    $CERT_DIR/" openssl x509 -in 
    "$CERT_DIR/server.crt" -noout 
    -dates
    ;;
  # ──────────────────────────────────────────────── 
  # Let's Encrypt — para produção 
  # com domínio real 
  # ────────────────────────────────────────────────
  letsencrypt) DOMAIN="${2:-}" 
    EMAIL="${3:-}" if [[ -z 
    "$DOMAIN" || -z "$EMAIL" ]]; 
    then
      echo "❌ Uso: $0 letsencrypt 
      <dominio> <email>" echo " 
      Ex: $0 letsencrypt 
      api.meuprojeto.com.br 
      admin@meuprojeto.com.br" 
      exit 1
    fi echo "🌐 Solicitando 
    certificado Let's Encrypt para 
    $DOMAIN..." if ! command -v 
    certbot &>/dev/null; then
      sudo apt-get install -y 
      certbot
    fi sudo certbot certonly 
    --standalone \
      --non-interactive \ 
      --agree-tos \ --email 
      "$EMAIL" \ -d "$DOMAIN"
    # Copia para o diretório do 
    # projeto
    sudo cp 
    "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" 
    "$CERT_DIR/server.crt" sudo cp 
    "/etc/letsencrypt/live/$DOMAIN/privkey.pem" 
    "$CERT_DIR/server.key" sudo 
    chown "$USER:$USER" 
    "$CERT_DIR"/*.{crt,key} echo 
    "✅ Certificado Let's Encrypt 
    instalado para $DOMAIN" echo " 
    Renovação automática: sudo 
    certbot renew"
    ;;
  *) echo "❌ Modo inválido. Use: 
    self-signed | letsencrypt" 
    exit 1
    ;;
esac
# Ajusta permissões
chmod 600 "$CERT_DIR/server.key" 
chmod 644 "$CERT_DIR/server.crt" 
echo "" echo "📋 Próximos passos:" 
echo " docker compose up -d nginx" 
echo " curl -k 
https://localhost/health"
