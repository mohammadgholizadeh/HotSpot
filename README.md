# HotSpot Backend Service

HotSpot is a Go service that queues trip requests with RabbitMQ, indexes locations with Uber H3, and caches hotspots in Redis.

## Stack overview

- **API**: Echo HTTP server exposing `/api/v1` routes.
- **Queues**: RabbitMQ exchange `rides` with `intra_city` / `inter_city` queues.
- **Storage**: PostgreSQL via storage layer in `internal/storage/`.
- **Caching**: Redis sorted sets store hotspot counts (3‑hour TTL).
- **Geospatial**: H3 resolution 8 converts lat/long pairs to hex cells.

## Configuration

The app loads YAML via `CONFIG_PATH` (defaults to `configs/config.yaml`). Important fields:

```yaml
server.port
postgres.host / user / password / db
redis.addr / password / db
broker.url  
hotspot.resolution / threshold / decay_half_life_min
```

For containers use `configs/docker.yaml`, which already maps hosts to compose/Helm service names.

## Run locally

- **API**
  ```bash
  go run cmd/app/main.go serve
  ```
- **Worker**
  ```bash
  go run cmd/app/main.go worker
  ```
- **Everything (Docker)**
  ```bash
  docker-compose up -d
  ```
