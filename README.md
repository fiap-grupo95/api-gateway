# API Gateway

Porta de entrada única do sistema FIAP Secure Systems. Responsável por autenticação, validação de requisições e roteamento seguro para os serviços internos.

---

## Descrição do Problema

Em uma arquitetura de microsserviços, expor cada serviço diretamente à internet cria uma superfície de ataque distribuída e dificulta a aplicação consistente de políticas de segurança (autenticação, rate limiting, validação de entrada). Um API Gateway centraliza essas responsabilidades, garantindo que nenhuma requisição mal-formada ou não autenticada chegue aos serviços internos.

**Desafios específicos endereçados:**

- Validar o tipo real do arquivo por conteúdo binário (não pela extensão, que pode ser forjada)
- Impedir abusos com rate limiting por IP sem depender de estado externo
- Propagar rastreamento de requisições (`X-Request-ID`) entre todos os serviços
- Proteger rotas com JWT sem expor o secret aos serviços downstream

---

## Arquitetura Proposta

```
Cliente
  │
  ▼
┌─────────────────────────────────────────────────────┐
│                    API Gateway :8080                 │
│                                                     │
│  gin.Recovery → New Relic → RequestID → Logger      │
│       → RateLimiter (100 req/min/IP) → JWTAuth      │
│                                                     │
│  ┌──────────────┐   ┌──────────────────────────┐   │
│  │ /api/diagrams│   │ /api/process/:id/status  │   │
│  │ /api/reports/│   │ client/orchestrator.go   │   │
│  │ :id          │   │ client/report.go         │   │
│  └──────────────┘   └──────────────────────────┘   │
└─────────────────────────────────────────────────────┘
         │                          │
         ▼                          ▼
  upload-orchestrator          report-service
       :8081                       :8083
```

### Camadas internas (Clean Architecture)

```
internal/
├── config/         ← JWT_SECRET (≥32 chars), MIME types, max upload size
├── middleware/
│   ├── auth.go         ← JWT HS256, WithValidMethods(["HS256"])
│   ├── request_id.go   ← gera/propaga X-Request-ID
│   ├── logger.go       ← log estruturado por request (zap)
│   └── rate_limiter.go ← golang.org/x/time/rate, 100 req/min, burst 20
├── client/
│   ├── orchestrator.go ← Upload (Content-Length explícito) + GetStatus
│   └── report.go       ← GetReport
└── handler/
    └── diagram.go      ← MIME sniff 512 bytes, sanitizeFilename, mapUpstreamStatus
```

### Decisões de segurança

| Controle | Implementação |
|---|---|
| Autenticação | JWT HS256, secret mínimo 32 chars, `WithValidMethods` fixado |
| Validação de tipo | `http.DetectContentType` nos primeiros 512 bytes (conteúdo binário) |
| Limitação de tamanho | `http.MaxBytesReader` antes de qualquer parse de multipart |
| Rate limiting | `golang.org/x/time/rate` por IP, burst 20, limpeza de entradas inativas |
| Rastreamento | `X-Request-ID` gerado (UUID) ou propagado do cliente |
| Isolamento de erros | `mapUpstreamStatus` — não vaza detalhes internos para o cliente |

---

## Fluxo da Solução

```
POST /api/diagrams
  1. RateLimiter verifica tokens disponíveis para o IP
  2. JWTAuth valida Bearer token (HS256, exp obrigatório)
  3. MaxBytesReader limita body a MAX_UPLOAD_SIZE_MB
  4. FormFile lê campo "diagram"
  5. DetectContentType nos primeiros 512 bytes
  6. Rejeita se MIME não está em ALLOWED_MIME_TYPES → 415
  7. Lê restante do arquivo em memória
  8. OrchestratorClient.Upload() com Content-Length explícito
  9. Propaga X-Request-ID e X-Filename (sanitizado)
 10. Retorna 202 com {process_id, status, created_at}

GET /api/process/:processId/status
  1. JWTAuth
  2. OrchestratorClient.GetStatus() com X-Request-ID
  3. mapUpstreamStatus propaga 404/400, retorna 502 para demais erros

GET /api/reports/:reportId
  1. JWTAuth
  2. ReportClient.GetReport() com X-Request-ID
  3. mapUpstreamStatus propaga 404/400
```

---

## Instruções de Execução

### Variáveis de ambiente

| Variável | Obrigatório | Padrão | Descrição |
|---|---|---|---|
| `JWT_SECRET` | Sim | — | Secret HS256 (mínimo 32 caracteres) |
| `UPLOAD_ORCHESTRATOR_URL` | Sim | — | Ex: `http://upload-orchestrator:8081` |
| `REPORT_SERVICE_URL` | Sim | — | Ex: `http://report-service:8083` |
| `PORT` | Não | `8080` | Porta HTTP |
| `MAX_UPLOAD_SIZE_MB` | Não | `10` | Limite de upload em MB (máx: 50) |
| `ALLOWED_MIME_TYPES` | Não | `image/png,image/jpeg,application/pdf` | MIME types aceitos |
| `NEW_RELIC_LICENSE_KEY` | Não | — | Chave New Relic (40 chars) |
| `NEW_RELIC_APP_NAME` | Não | — | Nome da aplicação no New Relic |

### Executar com Docker Compose (recomendado)

```bash
# A partir da raiz do projeto
cp .env.example .env   # edite JWT_SECRET e LLM_API_KEY
docker compose up --build -d api-gateway

# Verificar saúde
curl http://localhost:8080/ping
```

### Executar localmente (desenvolvimento)

```bash
cd api-gateway

# Instalar dependências
go mod download

# Configurar variáveis (ajuste conforme necessário)
export JWT_SECRET="um-secret-de-pelo-menos-32-caracteres-aqui"
export UPLOAD_ORCHESTRATOR_URL="http://localhost:8081"
export REPORT_SERVICE_URL="http://localhost:8083"

# Executar
go run main.go
```

### Gerar token JWT para testes

```bash
# A partir da raiz do projeto
source .env
go run api-gateway/docs/generate_token.go
```

### Endpoints

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| `GET` | `/ping` | Não | Healthcheck |
| `POST` | `/api/diagrams` | JWT | Upload de diagrama |
| `GET` | `/api/process/:processId/status` | JWT | Status do processamento |
| `GET` | `/api/reports/:reportId` | JWT | Relatório de análise |

---

## Segurança

### Requisitos básicos adotados

| Controle | Implementação |
|---|---|
| Autenticação | JWT HS256, `WithValidMethods(["HS256"])` — algoritmo fixado para prevenir algorithm confusion |
| Validação de secret | `config.Load()` rejeita `JWT_SECRET` com menos de 32 caracteres na inicialização |
| Detecção de MIME | `http.DetectContentType` nos primeiros 512 bytes do conteúdo binário — não confia na extensão |
| Limite de tamanho | `http.MaxBytesReader` aplicado antes de qualquer parse de multipart |
| Rate limiting | `golang.org/x/time/rate` por IP: 100 req/min, burst 20, limpeza automática de entradas inativas |
| Rastreamento | `X-Request-ID` gerado (UUID) ou preservado do cliente; propagado a todos os serviços downstream |
| Isolamento de erros | `mapUpstreamStatus` — nunca vaza mensagens de erro internas para o cliente externo |

### Validação de entradas não confiáveis

- **Tipo de arquivo:** MIME detectado por conteúdo binário (`http.DetectContentType`), não pela extensão ou pelo header `Content-Type` enviado pelo cliente.
- **Tamanho:** `http.MaxBytesReader(w, body, maxBytes)` limita o body antes de qualquer leitura — impede ataques de payload excessivo.
- **Nome do arquivo:** `sanitizeFilename` usa `filepath.Base` + remoção de `..`, `/` e `\` — previne path traversal (CWE-22).
- **Parâmetros de rota:** `processId` e `reportId` são passados diretamente ao upstream; a validação UUID ocorre no `upload-orchestrator` e `report-service`.
- **JWT:** apenas tokens HS256 com `exp` válido são aceitos; `AbortWithStatusJSON` interrompe a cadeia de middleware imediatamente.

### Comunicação entre serviços

- URLs dos serviços internos (`UPLOAD_ORCHESTRATOR_URL`, `REPORT_SERVICE_URL`) configuradas exclusivamente por variáveis de ambiente — sem SSRF por entrada do usuário.
- `X-Request-ID` propagado em todos os requests HTTP internos para rastreabilidade de ponta a ponta.
- Timeouts do `http.Client`: 60s para upload (arquivo pode ser grande), 15s para consultas.
- Erros do upstream são mapeados para códigos HTTP semânticos sem expor detalhes internos ao cliente.

### Principais riscos e limitações

| Risco | Severidade | Mitigação atual | Recomendação para produção |
|---|---|---|---|
| Sem revogação de JWT | Média | Expiração de 24h no token de teste | Implementar blocklist com Redis ou reduzir TTL |
| Sem mTLS para serviços upstream | Média | Rede Docker isolada | TLS mútuo ou service mesh |
| Rate limiter em memória | Baixa | Reiniciar o gateway reseta contadores | Usar Redis para estado distribuído em multi-instância |
| CORS não configurado | Baixa | Sem frontend SPA — apenas API | Configurar `Access-Control-Allow-Origin` se necessário |

---

### Build e testes

```bash
cd api-gateway
go build ./...
go vet ./...
```
