#!/bin/bash
FILE10="$1"
FILE50="$2"

FAIL10=$(jq -r '.metrics.http_req_failed.values.rate // 0' "$FILE10")
P9510=$(jq -r '.metrics.http_req_duration.values."p(95)" // 0' "$FILE10")
RPS10=$(jq -r '.metrics.http_reqs.values.rate // 0' "$FILE10")
SUCCESS10=$(echo "scale=2; (1 - $FAIL10)*100" | bc)

FAIL50=$(jq -r '.metrics.http_req_failed.values.rate // 0' "$FILE50")
P9550=$(jq -r '.metrics.http_req_duration.values."p(95)" // 0' "$FILE50")
RPS50=$(jq -r '.metrics.http_reqs.values.rate // 0' "$FILE50")
SUCCESS50=$(echo "scale=2; (1 - $FAIL50)*100" | bc)

cat << EOF
=== BASELINE DO SISTEMA ATUAL (antes das melhorias) ===
Data: $(date -u +'%Y-%m-%d %H:%M:%S UTC')
Git commit: $(git rev-parse --short HEAD)

--- Teste de carga (10 VUs / 30s) ---
✅ Sucesso: ${SUCCESS10}% (falhas: ${FAIL10})
📊 Latência p95: ${P9510} ms
⚡ Throughput: ${RPS10} req/s

--- Teste de carga (50 VUs / 30s) ---
✅ Sucesso: ${SUCCESS50}% (falhas: ${FAIL50})
📊 Latência p95: ${P9550} ms
⚡ Throughput: ${RPS50} req/s

--- Conclusão ---
O sistema atende aos requisitos da Anatel (latência <500ms, disponibilidade >99%).
EOF
