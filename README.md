# 🚀 LogMesh Web Application

**LogMesh** is a distributed log aggregation and monitoring platform built with **Go**, **Gin**, and **Next.js**.

It is designed as a serious backend portfolio project: a lightweight, from-scratch learning version of systems like Fluent Bit, OpenSearch, and Kibana.

The current version gives you a working local platform with:

- ⚡ Fast structured log ingestion
- 📦 Bulk log ingestion
- 🧾 Plain-text log parsing
- 🔍 Search and filtering
- 📊 Analytics summaries
- 🧠 Runtime metrics
- 🔐 API key management
- 🚦 In-memory rate limiting
- 📡 Live log streaming with Server-Sent Events
- 📤 CSV export
- 🖥️ Next.js dashboard

> Current storage is intentionally in-memory. This keeps the first version simple while the API contract, dashboard, and domain model stabilize before adding Kafka, OpenSearch, PostgreSQL, and Redis.

## 🧱 Architecture

```text
Applications
    |
    | POST /v1/logs
    v
Go API / Collector
    |
    | validate, normalize, mask sensitive fields
    v
In-memory log store
    |
    +--> Search API
    +--> Analytics API
    +--> CSV Export
    +--> SSE Live Stream
    +--> Next.js Dashboard
```

Future architecture:

```text
Applications
    |
    v
Go Collector
    |
    v
Kafka
    |
    v
Go Processor Worker Pool
    |
    v
OpenSearch
    |
    v
Search API + Analytics API
    |
    v
Next.js Dashboard

PostgreSQL -> users, projects, API keys
Redis      -> rate limits, cache
Prometheus -> metrics
Grafana    -> monitoring
```

## 🛠️ Tech Stack

- **Backend:** Go, Gin
- **Frontend:** Next.js, TypeScript, React, Recharts, Lucide icons
- **Storage today:** In-memory store
- **Planned storage:** Kafka, OpenSearch, PostgreSQL, Redis
- **Dev tooling:** Docker Compose scaffold, Go tests, Next production build

## 📁 Project Structure

```text
LogMesh/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── collector/
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   └── service/
├── dashboard/
│   ├── app/
│   └── lib/
├── .env
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## ⚙️ Environment Variables

Backend `.env`:

```env
LOGMESH_ENV=development
LOGMESH_HTTP_ADDR=:8081
LOGMESH_LOG_LEVEL=debug
LOGMESH_MAX_STORED_LOGS=10000
LOGMESH_REQUIRE_API_KEY=false
LOGMESH_RATE_LIMIT_REQUESTS=120
LOGMESH_RATE_LIMIT_WINDOW_SECONDS=60
```

Frontend `dashboard/.env.local`:

```env
NEXT_PUBLIC_LOGMESH_API_URL=http://localhost:8081
```

## ▶️ Run Backend

```powershell
go mod tidy
go run ./cmd/api
```

Backend URL:

```text
http://localhost:8081
```

Health check:

```powershell
Invoke-RestMethod http://localhost:8081/healthz
```

## 🖥️ Run Dashboard

```powershell
cd dashboard
npm install
npm run dev
```

Frontend URL:

```text
http://localhost:3000
```

## 🔥 API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/healthz` | Health check |
| `POST` | `/v1/logs` | Ingest one structured log |
| `POST` | `/v1/logs/bulk` | Ingest up to 500 logs |
| `POST` | `/v1/logs/parse` | Parse and ingest a plain-text log |
| `GET` | `/v1/logs` | Search logs |
| `GET` | `/v1/logs/:id` | Get one log by ID |
| `GET` | `/v1/logs/export` | Export filtered logs as CSV |
| `GET` | `/v1/analytics` | Get analytics summary |
| `GET` | `/v1/sources` | Get discovered log sources |
| `GET` | `/v1/runtime` | Get runtime memory and process stats |
| `GET` | `/v1/stream/logs` | Stream live logs with SSE |
| `GET` | `/v1/api-keys` | List API keys |
| `POST` | `/v1/api-keys` | Create API key |
| `DELETE` | `/v1/api-keys/:id` | Revoke API key |

## 📨 Ingest One Log

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/v1/logs `
  -ContentType application/json `
  -Body '{
    "service": "payment-service",
    "environment": "production",
    "level": "ERROR",
    "message": "Payment processing failed",
    "host": "server-03",
    "trace_id": "abc123",
    "metadata": {
      "database": "payments",
      "access_token": "secret-token"
    }
  }'
```

Sensitive metadata fields are masked automatically:

```json
{
  "access_token": "[REDACTED]"
}
```

## 📦 Bulk Ingest

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/v1/logs/bulk `
  -ContentType application/json `
  -Body '{
    "logs": [
      {
        "service": "api-gateway",
        "environment": "development",
        "level": "INFO",
        "message": "Request completed"
      },
      {
        "service": "payment-service",
        "environment": "production",
        "level": "WARN",
        "message": "Gateway latency is high"
      }
    ]
  }'
```

## 🧾 Parse Plain-Text Logs

Supported format:

```text
2026-08-30 10:21:22 ERROR Payment failed
```

Request:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/v1/logs/parse `
  -ContentType application/json `
  -Body '{
    "service": "payment-service",
    "environment": "production",
    "host": "server-03",
    "trace_id": "trace-001",
    "line": "2026-08-30 10:21:22 ERROR Payment failed"
  }'
```

If the text line does not match the timestamp/level pattern, LogMesh stores it as an `INFO` message.

## 🔍 Search Logs

```powershell
Invoke-RestMethod "http://localhost:8081/v1/logs?service=payment-service&level=ERROR&search=payment"
```

Supported query parameters:

- `service`
- `environment`
- `level`
- `from`
- `to`
- `search`
- `trace_id`
- `host`
- `limit`
- `offset`

Example with time range:

```text
GET /v1/logs?level=ERROR&from=2026-08-30T10:00:00Z&to=2026-08-30T11:00:00Z
```

## 📤 Export CSV

```powershell
Invoke-WebRequest `
  -Uri "http://localhost:8081/v1/logs/export?level=ERROR&limit=500" `
  -OutFile logs.csv
```

CSV columns:

```text
id,timestamp,level,service,environment,host,trace_id,message
```

## 📊 Analytics

```powershell
Invoke-RestMethod http://localhost:8081/v1/analytics
```

Returns:

- Total logs
- Error count
- Warning count
- Error rate
- Service count
- Level counts
- Top services
- Top errors
- Timeline buckets

## 🧩 Sources

```powershell
Invoke-RestMethod http://localhost:8081/v1/sources
```

Sources are grouped by service and environment, with host count, log count, and last-seen time.

## 🔐 API Keys

Create an API key:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/v1/api-keys `
  -ContentType application/json `
  -Body '{"name":"Production collector"}'
```

List API keys:

```powershell
Invoke-RestMethod http://localhost:8081/v1/api-keys
```

Revoke an API key:

```powershell
Invoke-RestMethod `
  -Method Delete `
  -Uri http://localhost:8081/v1/api-keys/<api-key-id>
```

Enable required API-key ingestion:

```env
LOGMESH_REQUIRE_API_KEY=true
```

Then send logs with:

```http
X-API-Key: lm_live_xxxxxxxxx
```

or:

```http
Authorization: Bearer lm_live_xxxxxxxxx
```

Only API-key hashes are stored internally. The plaintext key is returned only once when it is created.

## 📡 Live Log Streaming

LogMesh exposes a Server-Sent Events stream:

```text
GET /v1/stream/logs
```

The dashboard uses this stream while Live mode is enabled. New logs appear without waiting for the next polling interval.

## 🧠 Runtime Stats

```powershell
Invoke-RestMethod http://localhost:8081/v1/runtime
```

Returns uptime, Go version, goroutine count, memory usage, GC count, and stored-log count.

## 🚦 Rate Limiting

Rate limiting is controlled by:

```env
LOGMESH_RATE_LIMIT_REQUESTS=120
LOGMESH_RATE_LIMIT_WINDOW_SECONDS=60
```

That means each client IP can make 120 requests per 60-second window by default.

When exceeded, the API returns:

```http
429 Too Many Requests
```

## 🖥️ Dashboard Features

The Next.js dashboard includes:

- 🔎 Search bar for log messages
- 🎚️ Filters for service, environment, and level
- 📋 Log table
- 🧾 Log detail panel
- ➕ Structured log ingest form
- 🧾 Raw text log parser
- 📦 Demo log seeding through bulk ingest
- 📊 Analytics charts
- 🧩 Source discovery table
- 🔐 API key creation and revocation
- 📤 CSV export button
- 📡 Live mode using SSE
- 🧠 Runtime stats in settings

## 🧪 Testing

Run backend tests:

```powershell
go test ./...
```

Build frontend:

```powershell
cd dashboard
npm run build
```

## 🐳 Docker

Docker files are included for future service orchestration:

- `Dockerfile`
- `docker-compose.yml`

Current compose services include:

- API
- PostgreSQL
- Redis
- OpenSearch

## 🧭 Roadmap

### ✅ Completed

- Go module
- Gin API
- Config loading
- `.env` loading
- Health endpoint
- Structured log ingestion
- Bulk ingestion
- Plain-text parsing
- In-memory search
- Sensitive metadata masking
- CSV export
- Analytics summary
- Source discovery
- Runtime metrics
- API key creation/revocation
- Optional API-key ingestion auth
- In-memory rate limiting
- SSE live log stream
- Next.js dashboard
- Frontend build verification
- Backend unit tests

### 🚧 Next Steps

1. Add Kafka producer for `POST /v1/logs`.
2. Add Kafka consumer and processor command.
3. Add bounded worker pool for backpressure.
4. Add OpenSearch indexing and search query translation.
5. Add PostgreSQL persistence for users, projects, sources, and API keys.
6. Add Redis-backed distributed rate limiting.
7. Add JWT dashboard authentication.
8. Add multi-tenant project isolation.
9. Add retry handling and dead-letter queue.
10. Add Prometheus metrics endpoint.
11. Add Grafana dashboard.
12. Add load-testing scripts.

## 🧠 Design Decisions

### Why Go?

Go is a strong fit for infrastructure software because it has goroutines, channels, fast HTTP servers, low runtime overhead, simple deployment, and a strong standard library.

### Why Kafka later?

Kafka will provide durable buffering, replay, partitions, consumer groups, and horizontal processor scaling.

### Why OpenSearch later?

Log search needs full-text search, filtering, aggregations, time-based indexes, and fast query performance over large volumes.

### Why PostgreSQL later?

PostgreSQL should own metadata such as users, projects, API keys, log sources, and configuration. High-volume log documents should go to OpenSearch, not PostgreSQL.



## ✨ Status

LogMesh currently has a working local backend and frontend. It is ready for the next major phase: replacing the in-memory store with Kafka plus OpenSearch.
