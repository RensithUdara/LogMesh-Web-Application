# LogMesh

Distributed log aggregation and monitoring platform, built in Go.

This repository starts with the smallest useful slice of the system:

- `POST /v1/logs` ingests structured logs.
- `GET /v1/logs` searches the in-memory store with filters.
- `GET /v1/logs/:id` returns one log event.
- `GET /v1/analytics` returns counts, error rate, top services, top errors, and timeline buckets.
- `GET /v1/sources` returns discovered service/environment sources.
- `GET /v1/stream/logs` streams newly ingested logs over Server-Sent Events.
- `GET /v1/api-keys`, `POST /v1/api-keys`, and `DELETE /v1/api-keys/:id` manage API keys.
- `GET /healthz` reports service health.
- `dashboard/` provides a Next.js dashboard for ingestion, search, analytics, sources, keys, and settings.

The current storage is intentionally in-memory so the API contract and domain model can stabilize before Kafka, OpenSearch, PostgreSQL, and Redis are wired in.

## Run Locally

```powershell
go mod tidy
go run ./cmd/api
```

The API listens on `http://localhost:8081` by default.

Set `LOGMESH_REQUIRE_API_KEY=true` to require `X-API-Key` or `Authorization: Bearer <key>` on log ingestion.

## Run Dashboard

```powershell
cd dashboard
npm install
npm run dev
```

The dashboard listens on `http://localhost:3000` by default and calls the Go API at `http://localhost:8081`.

To point it at another backend:

```powershell
$env:NEXT_PUBLIC_LOGMESH_API_URL="http://localhost:8081"
npm run dev
```

## Example Ingest

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

## Example Search

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

## Roadmap

1. Go API foundation with validation, config, logging, and tests.
2. Durable ingestion through Kafka producer and consumer packages.
3. Processor worker pool with bounded channels and graceful shutdown.
4. OpenSearch bulk indexing and query translation.
5. PostgreSQL metadata for projects, users, and API keys.
6. API key authentication and rate limiting.
7. Next.js dashboard with search, filtering, analytics, and live streaming.
8. Retry handling, dead-letter queue, metrics, Prometheus, and Grafana.

## Tests

```powershell
go test ./cmd/... ./internal/...
cd dashboard
npm run build
```
