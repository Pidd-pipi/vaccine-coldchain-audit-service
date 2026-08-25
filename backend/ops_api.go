package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// opsSeed returns the baseline vaccine-coldchain audit work records.
func opsSeed() []OpsRecord {
	return []OpsRecord{
		{ID: "batch-1001", Subject: "流感疫苗批次冷链复核", Owner: "auditor-lin", Status: OpsStatusQueued, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "东城接种门诊", "operator": "auditor-lin", "evidence": "temp-log-1"}, Revision: 1},
		{ID: "batch-1002", Subject: "儿童联合疫苗批次放行审核", Owner: "auditor-zhao", Status: OpsStatusActive, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "南湖社区门诊", "operator": "auditor-zhao", "evidence": "temp-log-2"}, Revision: 1},
		{ID: "batch-1003", Subject: "百白破疫苗批次隔离复查", Owner: "auditor-wu", Status: OpsStatusPaused, Priority: OpsPriorityCritical, Labels: map[string]string{"site": "西城接种点", "operator": "auditor-wu", "evidence": "temp-log-3"}, Revision: 1},
	}
}

// newOpsHandler exposes the operations workflow over HTTP.
func newOpsHandler(service *OpsService) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/api/ops/records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleOpsSearch(w, r, service)
		case http.MethodPost:
			handleOpsCreate(w, r, service)
		default:
			opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})
	m.HandleFunc("/api/ops/records/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/ops/records/")
		switch {
		case strings.HasSuffix(path, "/transition"):
			if r.Method != http.MethodPost {
				opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			handleOpsTransition(w, r, service, strings.TrimSuffix(path, "/transition"))
		default:
			if r.Method != http.MethodGet {
				opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			handleOpsGet(w, r, service, path)
		}
	})
	m.HandleFunc("/api/ops/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			opsJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		opsJSON(w, http.StatusOK, service.Snapshot())
	})
	return m
}

type opsCreateRequest struct {
	ID       string            `json:"id"`
	Subject  string            `json:"subject"`
	Owner    string            `json:"owner"`
	Status   OpsStatus         `json:"status"`
	Priority OpsPriority       `json:"priority"`
	Labels   map[string]string `json:"labels"`
}

type opsTransitionRequest struct {
	Expected int       `json:"expected"`
	Target   OpsStatus `json:"target"`
}

func handleOpsCreate(w http.ResponseWriter, r *http.Request, service *OpsService) {
	var req opsCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		opsJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	record := OpsRecord{ID: req.ID, Subject: req.Subject, Owner: req.Owner, Status: req.Status, Priority: req.Priority, Labels: req.Labels}
	if record.Status == "" {
		record.Status = OpsStatusQueued
	}
	created, err := service.Create(r.Context(), record)
	if err != nil {
		opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
		return
	}
	opsJSON(w, http.StatusCreated, created)
}

func handleOpsSearch(w http.ResponseWriter, r *http.Request, service *OpsService) {
	q := OpsQuery{
		Subject:  r.URL.Query().Get("subject"),
		Status:   OpsStatus(r.URL.Query().Get("status")),
		Priority: OpsPriority(r.URL.Query().Get("priority")),
		Owner:    r.URL.Query().Get("owner"),
	}
	if v := r.URL.Query().Get("page"); v != "" {
		q.Page, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		q.PageSize, _ = strconv.Atoi(v)
	}
	page, err := service.Search(r.Context(), q)
	if err != nil {
		opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
		return
	}
	opsJSON(w, http.StatusOK, page)
}

func handleOpsGet(w http.ResponseWriter, r *http.Request, service *OpsService, id string) {
	record, err := service.Get(r.Context(), id)
	if err != nil {
		opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
		return
	}
	opsJSON(w, http.StatusOK, record)
}

func handleOpsTransition(w http.ResponseWriter, r *http.Request, service *OpsService, id string) {
	var req opsTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		opsJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	record, err := service.Transition(r.Context(), id, req.Expected, req.Target, opsActorFromRequest(r))
	if err != nil {
		opsJSON(w, opsHTTPStatus(err), map[string]string{"error": err.Error()})
		return
	}
	opsJSON(w, http.StatusOK, record)
}

// opsHTTPStatus maps an operations error to an HTTP status code. A revision
// conflict (e.g. a duplicate audit record submission) is 409 Conflict; an
// illegal status transition is 422 Unprocessable Entity — the request is
// well-formed but cannot be applied to the current record state. These were
// previously collapsed to 500, which is why the audit tool saw server errors
// instead of the correct client-error classifications.
func opsHTTPStatus(err error) int {
	switch opsCode(err) {
	case "not_found":
		return http.StatusNotFound
	case "conflict":
		return http.StatusConflict
	case "invalid":
		return http.StatusBadRequest
	case "transition":
		return http.StatusUnprocessableEntity
	case "policy":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
