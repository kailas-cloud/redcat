# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

RedCat is a geospatial POI (Points of Interest) search system that uses vector similarity search for category matching. It combines a Python embedding service with a Go API server, backed by Redis vector sets.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│  API (Go)   │────▶│    Redis    │
└─────────────┘     └──────┬──────┘     │  (Vector)   │
                          │            └─────────────┘
                          ▼
                   ┌─────────────┐
                   │  Embedder   │
                   │  (Python)   │
                   └─────────────┘
```

### Components

**`archive/redcat/` - Go API Server**
- Standard Go layout: `cmd/redcat/main.go` entry point
- `internal/http/api/` - HTTP handlers for `/categories` and `/places`
- `internal/service/` - Business logic for categories and places search
- `internal/storage/` - Redis storage layer using rueidis client
- `internal/clients/embedder/` - HTTP client for embedder service
- Uses Redis VSIM command for vector similarity search

**`archive/embedder/` - Python Embedding Service**
- `app.py` - FastAPI service serving `/embed` endpoint
- Uses ONNX runtime with e5-multilingual-large model
- `embedder.py` - Batch embedding script for category data

**`archive/migrations/` - Data Migration Scripts**
- `categories.py` - Loads categories with embeddings into Redis vector set (VADD/VSETATTR)
- `places.py` - Loads POI data into Redis hashes

**`archive/balancer/` - Geospatial Utilities**
- H3 hexagon labeling for POIs at resolutions 5-8
- Aggregation scripts for density analysis

## Commands

All commands run from `archive/` directory using [just](https://github.com/casey/just):

```bash
# Docker operations
just up          # Build and start all services
just down        # Stop services
just logs        # Follow container logs
just clean       # Stop and remove volumes

# Data operations (requires .venv)
just migrate     # Run categories.py and places.py migrations
just embedder    # Generate category embeddings
just balancer    # Label and aggregate H3 hexagons

# Utilities
just redis       # Open redis-cli in container
just insight     # Open RedisInsight UI (localhost:5540)
just kepler      # Open Kepler.gl UI (localhost:8080)
```

### Go Development

```bash
cd archive/redcat
go build ./cmd/redcat
go test ./...
```

### Python Development

```bash
cd archive
python -m venv .venv
source .venv/bin/activate
pip install -r embedder/requirements.txt
pip install redis pandas fastembed h3  # for migrations/balancer
```

## Environment Variables

Key variables (defined in `.env`):
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_DB`, `REDIS_PASSWORD` - Redis connection
- `EMBEDDER_MODEL`, `EMBEDDER_MODEL_PATH`, `EMBEDDER_OUT` - Embedding model config
- `FS_CATEGORIES_RAW_DATA` - Path to raw category parquet file
- `POIS_CSV` - Path to POI CSV data
- `MAPBOX_TOKEN` - For Kepler.gl visualization

## Service Ports

- `6379` - Redis
- `5540` - RedisInsight
- `8080` - Kepler.gl
- `8081` - Embedder API
- `8082` - RedCat Go API
- `8083` - Static data server (nginx)

## Data Flow

1. Category embeddings generated via `embedder.py` → JSON file
2. `migrations/categories.py` loads embeddings into Redis vector set
3. POI data loaded via `migrations/places.py` into Redis hashes
4. API queries: text → embedder service → vector → Redis VSIM → results
