package httpapi

import (
	"encoding/json"
	"errors"
	"example.com/vaccine-coldchain-audit-service/domain"
	"example.com/vaccine-coldchain-audit-service/store"
	"example.com/vaccine-coldchain-audit-service/validation"
	"net/http"
	"strings"
)

type Handler struct{ Store *store.Store }

func New(s *store.Store) *Handler { return &Handler{Store: s} }
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && r.URL.Path == "/api/v1/batches" {
		write(w, 200, map[string]any{"items": h.Store.List()})
		return
	}
	if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v1/batches/") {
		h.update(w, r)
		return
	}
	http.Error(w, "route not found", 404)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) != 5 || p[0] != "api" || p[1] != "v1" || p[2] != "batches" || p[4] != "status" || p[3] == "" {
		http.Error(w, "route not found", 404)
		return
	}
	var c domain.StatusChange
	if json.NewDecoder(r.Body).Decode(&c) != nil {
		http.Error(w, "invalid JSON body", 400)
		return
	}
	if e := validation.Status(c.Status); e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	item, e := h.Store.UpdateStatus(p[3], c.Status)
	if errors.Is(e, store.ErrNotFound) {
		http.Error(w, "batch not found", 404)
		return
	}
	write(w, 200, item)
}
func write(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
