package main

import (
	"example.com/vaccine-coldchain-audit-service/config"
	"example.com/vaccine-coldchain-audit-service/health"
	"example.com/vaccine-coldchain-audit-service/httpapi"
	"example.com/vaccine-coldchain-audit-service/store"
	"example.com/vaccine-coldchain-audit-service/web"
	"log"
	"net/http"
	"time"
)

func main() {
	c := config.Load()
	m := http.NewServeMux()
	m.HandleFunc("/healthz", health.Handler)
	m.Handle("/api/v1/", httpapi.New(store.New()))
	opsSvc := newOpsService(opsSeed())
	m.Handle("/api/ops/", opsEnterpriseMiddleware(newOpsHandler(opsSvc)))
	startAuditPruner(opsSvc.audit, time.Minute, make(chan struct{}))
	m.HandleFunc("/", web.Handler)
	log.Printf("vaccine coldchain audit listening on %s", c.Address())
	log.Fatal(serveAddress(c.Address(), m))
}

// startAuditPruner runs a background goroutine that periodically trims old
// audit events so the in-memory log stays bounded. It stops when stop closes.
func startAuditPruner(audit *OpsAudit, interval time.Duration, stop <-chan struct{}) {
	go runAuditPruner(audit, interval, stop)
}

// runAuditPruner is the blocking pruner loop; it returns once stop closes.
func runAuditPruner(audit *OpsAudit, interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			threshold := time.Now().Add(-24 * time.Hour)
			_, _ = audit.Prune(threshold)
		}
	}
}
