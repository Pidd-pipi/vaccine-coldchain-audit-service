# Vaccine Coldchain Audit Service

疫苗批次冷链审计 HTTP 服务，默认监听 `8080`，可由 `PORT` 环境变量调整。

## Layout

```text
.
├── backend/                 # Go module, application, tests, static assets, Dockerfile
├── database/                # persistence extension point
├── output/                  # verification record
├── prompt.txt               # task prompt
├── runtime_smoke.json       # startup contract
├── .env.example
└── .gitignore
```

## API and Health

- `GET /healthz` - health check
- `GET /api/v1/batches` - list vaccine batches
- `POST /api/v1/batches/{id}/status` - update a batch status
- `GET /` - embedded operator page

Supported statuses are `received`, `cold`, `quarantine`, `released`, and `recalled`.

## Run

```bash
cd backend
go test ./...
go build ./...
go run .
```
