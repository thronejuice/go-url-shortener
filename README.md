# go-url-shortener

URL shortener API in Go with a production-style pipeline: PostgreSQL, Redis, Docker, GitHub Actions, metrics.

## Quick start

```bash
make run          # http://localhost:8080/healthz
make test
make up           # app + postgres + redis via docker compose
```

## Layout

```
cmd/server        entrypoint
internal/config   env-based config
internal/server   routes + middleware
migrations        SQL migrations
```

## Roadmap

- [x] Week 1: project skeleton, `/healthz`, slog, graceful shutdown, CI
- [ ] Week 2: create/redirect links, PostgreSQL, migrations
- [ ] Week 3: tests (testcontainers), Redis cache, rate limiting
- [ ] Week 4: Docker/CI polish, push image to GHCR
- [ ] Week 5: Prometheus metrics + Grafana
- [ ] Week 6: deploy, architecture diagram, live demo link
