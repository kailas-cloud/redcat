# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

RedCat is a geospatial POI (Points of Interest) search API using Valkey (Redis-compatible) with FT.SEARCH for vector-based nearest neighbor queries.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API (Go)   │────▶│   Valkey    │
│             │     │   Fiber     │     │  FT.SEARCH  │
└─────────────┘     └─────────────┘     └─────────────┘
```

**Production URL:** https://redcat.kailas.cloud/api/v1/

### Components

**Go API Server (root)**
- `cmd/redcat/main.go` - Entry point
- `internal/api/` - HTTP handlers (Fiber) with structured JSON logging
- `internal/service/places/` - Business logic for places CRUD and search
- `internal/storage/valkey/` - Valkey storage layer using rueidis client
- `internal/config/` - Environment configuration
- `internal/domain/model/` - Domain models

**Deployment**
- `Dockerfile` - Multi-stage Go build for linux/amd64
- `k8s/` - Kubernetes manifests (namespace, deployment, service, ingress)
- `.github/workflows/deploy.yml` - CI/CD: auto-tag → build → push GHCR → deploy k8s

**Legacy (archive/)**
- Python embedder service, migrations, and balancer scripts (not used in current deployment)

## Commands

### Go Development

```bash
# Build
go build ./cmd/redcat

# Run unit tests
go test ./...

# Run integration tests (hits live API)
go test -tags=integration ./tests/integration/...

# Run integration tests against custom URL
REDCAT_API_URL=http://localhost:8080 go test -tags=integration ./tests/integration/...
```

### Docker

```bash
# Build image
docker build --platform linux/amd64 -t redcat .

# Run locally
docker run -p 8080:8080 \
  -e VALKEY_ADDRS=host.docker.internal:6379 \
  redcat
```

### Kubernetes

```bash
# View logs
kubectl logs -n redcat -l app=api -f

# Restart deployment
kubectl rollout restart deployment/api -n redcat

# Check status
kubectl get pods -n redcat
```

## Environment Variables

- `HTTP_ADDR` - Listen address (default `:8080`)
- `VALKEY_ADDRS` - Comma-separated Valkey addresses (default `localhost:6379`)
- `VALKEY_USER` - Valkey username (optional)
- `VALKEY_PASS` - Valkey password (optional)
- `VALKEY_INDEX` - FT.SEARCH index name (default `index_places`)
- `VALKEY_PREFIX` - Key prefix for places (default `places:`)

## API Endpoints

- `GET /healthz` - Health check (internal only, not exposed via ingress)
- `POST /api/v1/places` - Create place
- `GET /api/v1/places/:id` - Get place by ID
- `DELETE /api/v1/places/:id` - Delete place
- `POST /api/v1/places/search` - Search nearby places

### Search Request Example

```json
{
  "location": {"lat": 55.7558, "lon": 37.6173},
  "category_ids": ["13000"],
  "limit": 10
}
```

## Logging

API uses structured JSON logging via `log/slog`:
- Request logging middleware (method, path, status, latency, IP)
- Operation logging (search params, results count, errors)

View logs: `kubectl logs -n redcat -l app=api -f`

## Testing

**Unit tests:** `go test ./...`

**Integration tests:** `go test -tags=integration ./tests/integration/...`

Integration tests hit the live API at https://redcat.kailas.cloud/api/v1/ and test:
- Health check
- Place CRUD (create, get, delete)
- Search (nearby, with category filter)
- Validation (missing fields, invalid JSON)

Set `REDCAT_API_URL` to test against a different environment.

## CI/CD

On push to `main`:
1. Compute next semver tag (v0.0.X)
2. Build Docker image for linux/amd64
3. Push to ghcr.io/kailas-cloud/redcat
4. Create GitHub Release
5. Deploy to Kubernetes via kustomize

---

## Valkey Search (current path, no-ML)

- Index schema (idempotent via FT.INFO → FT.CREATE):
  - `FT.CREATE index_places ON HASH PREFIX 1 places: SCHEMA`
    - `category_ids TAG SEPARATOR ","`
    - `location VECTOR FLAT TYPE FLOAT32 DIM 3 DISTANCE_METRIC L2`
- Documents: `HSET places:{fsq_place_id}` with fields:
  - `id,name,lat,lon,address,category_ids,location`
  - `location` — 3×float32 (little-endian) вектор ECEF на единичной сфере из (lat, lon)
- Query builder rules (RediSearch syntax is strict):
  - AND — пробел между частями; OR — `|` в скобках
  - Категории: `@category_ids:{id1|id2|...}` (значения разделены запятой в HASH)
  - KNN секция: `=>[KNN {k} @location $vec]`
  - Параметры: `PARAMS 2 vec <binary_3xfloat32_le>`; всегда `DIALECT 2`
  - Возврат/лимиты: `RETURN 4 id name lat lon | LIMIT 0 {k}`
- Go helpers (рекомендации):
  - `CreateIndex(ctx)` — проверяет `FT.INFO`, затем `FT.CREATE`; игнорирует "Index already exists"
  - `buildKnnQuery(k, categoryIDs)` — собирает фильтр категорий + `=>[KNN ...]`
  - `SearchNearest` — строит `FT.SEARCH` через rueidis builder и `Dialect(2)`
- Поведение поиска:
  - Всегда top‑K ближайших по L2 над ECEF-векторами; радиус не используется.
  - H3 не используется в этой реализации.
