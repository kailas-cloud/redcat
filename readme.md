# RedCat

Geospatial POI search API using Valkey (Redis-compatible) with FT.SEARCH for vector-based KNN queries.

## Status: ✅ MVP Ready

**Production URL:** https://redcat.kailas.cloud/api/v1/

## Progress

### ✅ Completed
- [x] Go API server (Fiber) with CRUD operations for places
- [x] Valkey cluster integration (3 shards) via rueidis client
- [x] Vector search (KNN) using ECEF coordinates on unit sphere
- [x] Category filtering via TAG fields
- [x] Kubernetes deployment with auto-scaling
- [x] CI/CD: GitHub Actions → GHCR → k8s
- [x] Integration tests (search ranking, CRUD, validation)
- [x] **Verified: data distribution across shards is uniform**
- [x] **Verified: coordinator correctly aggregates FT.SEARCH results from all cluster nodes**

### 🔜 Next Steps
- [ ] Load 100M Foursquare POIs
- [ ] Performance benchmarks under load
- [ ] Pagination support
- [ ] Geo-radius search (alternative to KNN)

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/places` | Create place |
| GET | `/api/v1/places/:id` | Get place |
| DELETE | `/api/v1/places/:id` | Delete place |
| POST | `/api/v1/places/search` | Search nearby |

### Search Example
```bash
curl -X POST https://redcat.kailas.cloud/api/v1/places/search \
  -H "Content-Type: application/json" \
  -d '{"location":{"lat":34.7575,"lon":32.407},"limit":10}'
```

## Architecture

```
Client → API (Go/Fiber) → Valkey Cluster (3 shards)
                              ↓
                         FT.SEARCH KNN
                         (ECEF vectors)
```

## Data Schema (Foursquare)

| Field | Description |
|-------|-------------|
| fsq_place_id | Unique identifier |
| name | Place name |
| latitude, longitude | Coordinates |
| address, country | Location |
| fsq_category_ids | Category codes |
| tel, website, email, instagram | Contacts |
