#!/bin/bash
# Gera token JWT para teste (use a mesma secret definida no gateway)
JWT_SECRET="${JWT_SECRET:-changeme-in-production}"
# Usando curl com jq ou python para gerar token
python3 -c "
import jwt
import time
payload = {
    'user_id': 'testuser',
    'scopes': ['credit:write'],
    'exp': int(time.time()) + 86400,
    'iat': int(time.time()),
    'iss': 'realtime-credit-validator',
    'aud': 'anatel-gateway'
}
token = jwt.encode(payload, '$JWT_SECRET', algorithm='HS256')
print(token)
"
