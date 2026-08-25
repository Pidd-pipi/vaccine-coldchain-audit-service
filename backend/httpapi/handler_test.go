package httpapi

import (
	"bytes"
	"example.com/vaccine-coldchain-audit-service/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchAPI(t *testing.T) {
	s := httptest.NewServer(New(store.New()))
	defer s.Close()
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{{"GET", "/api/v1/batches", "", 200}, {"POST", "/api/v1/batches/vb-401/status", `{"status":"released"}`, 200}, {"POST", "/api/v1/batches/vb-401/status", `{"status":"warm"}`, 400}, {"POST", "/api/v1/batches/nope/status", `{"status":"recalled"}`, 404}} {
		req, _ := http.NewRequest(tc.method, s.URL+tc.path, bytes.NewBufferString(tc.body))
		res, e := http.DefaultClient.Do(req)
		if e != nil || res.StatusCode != tc.want {
			t.Fatalf("%s %s: err=%v status=%d", tc.method, tc.path, e, res.StatusCode)
		}
		res.Body.Close()
	}
}
