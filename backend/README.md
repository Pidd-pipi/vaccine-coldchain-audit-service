# Backend

This directory is the Go module `example.com/vaccine-coldchain-audit-service`.
The application entry point is `main.go`; embedded web assets are under `web/`.

```bash
go test ./...
go build ./...
go run .
```

The service listens on port `8080` by default and supports `PORT`.
