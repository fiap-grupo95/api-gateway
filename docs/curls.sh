#!/usr/bin/env bash
# =============================================================================
# API Gateway — Documentação de Endpoints
# Base URL: http://localhost:8080
# =============================================================================
#
# PRÉ-REQUISITO: gerar um JWT de teste
#   cd /caminho/do/projeto
#   source .env
#   JWT=$(go run api-gateway/docs/generate_token.go)
#
# =============================================================================

BASE_URL="http://localhost:8080"
JWT="${JWT_TOKEN:-SEU_TOKEN_AQUI}"
PROCESS_ID="${PROCESS_ID:-SEU_PROCESS_ID_AQUI}"
REPORT_ID="${REPORT_ID:-SEU_REPORT_ID_AQUI}"

# ─── Health Check ─────────────────────────────────────────────────────────────
# Verifica se o serviço está no ar (não requer auth)
curl -i "${BASE_URL}/ping"

# Resposta esperada:
# HTTP/1.1 200 OK
# X-Request-Id: <uuid>
# pong

# ─── Upload de Diagrama ───────────────────────────────────────────────────────
# POST /api/diagrams
# Requer: Authorization Bearer JWT
# Body: multipart/form-data com campo "diagram"
# MIME aceitos: image/png, image/jpeg, image/webp, application/pdf
# Tamanho máximo: 10MB
#
curl -i -X POST "${BASE_URL}/api/diagrams" \
  -H "Authorization: Bearer ${JWT}" \
  -H "X-Request-ID: test-upload-001" \
  -F "diagram=@/caminho/para/diagrama.png"

# Resposta esperada (202 Accepted):
# {
#   "process_id": "550e8400-e29b-41d4-a716-446655440000",
#   "status": "RECEBIDO",
#   "created_at": "2026-05-16T22:00:00Z"
# }

# ─── Consulta de Status ───────────────────────────────────────────────────────
# GET /api/process/:processId/status
# Requer: Authorization Bearer JWT
# Status possíveis: RECEBIDO | EM_PROCESSAMENTO | ANALISADO | ERRO
#
curl -i "${BASE_URL}/api/process/${PROCESS_ID}/status" \
  -H "Authorization: Bearer ${JWT}" \
  -H "X-Request-ID: test-status-001"

# Resposta esperada (200 OK) — em processamento:
# {
#   "process_id": "550e8400-...",
#   "status": "EM_PROCESSAMENTO"
# }

# Resposta esperada (200 OK) — concluído:
# {
#   "process_id": "550e8400-...",
#   "status": "ANALISADO",
#   "report_id": "660e8400-..."
# }

# ─── Leitura do Relatório ─────────────────────────────────────────────────────
# GET /api/reports/:reportId
# Requer: Authorization Bearer JWT
# Disponível apenas quando status = "ANALISADO"
#
curl -i "${BASE_URL}/api/reports/${REPORT_ID}" \
  -H "Authorization: Bearer ${JWT}" \
  -H "X-Request-ID: test-report-001"

# Resposta esperada (200 OK):
# {
#   "report_id": "660e8400-...",
#   "process_id": "550e8400-...",
#   "components": ["API Gateway", "Upload Service", "PostgreSQL", ...],
#   "risks": ["Sem autenticação no banco X", "Porta 5432 exposta", ...],
#   "recommendations": ["Adicionar TLS", "Usar IAM roles", ...],
#   "created_at": "2026-05-16T22:05:00Z"
# }

# ─── Casos de Erro ────────────────────────────────────────────────────────────

# Sem token → 401
curl -i -X POST "${BASE_URL}/api/diagrams" \
  -F "diagram=@/caminho/para/diagrama.png"

# Token inválido → 401
curl -i "${BASE_URL}/api/process/qualquer/status" \
  -H "Authorization: Bearer token_invalido"

# MIME não permitido (arquivo .exe) → 415
curl -i -X POST "${BASE_URL}/api/diagrams" \
  -H "Authorization: Bearer ${JWT}" \
  -F "diagram=@/tmp/test.exe"

# processId inválido (não UUID) → 400 do upstream
curl -i "${BASE_URL}/api/process/nao-e-uuid/status" \
  -H "Authorization: Bearer ${JWT}"

# reportId não encontrado → 404
curl -i "${BASE_URL}/api/reports/00000000-0000-0000-0000-000000000000" \
  -H "Authorization: Bearer ${JWT}"

# Rate limit (>20 burst por IP) — execute em loop para acionar 429
for i in $(seq 1 25); do
  curl -s -o /dev/null -w "%{http_code}\n" "${BASE_URL}/ping"
done
