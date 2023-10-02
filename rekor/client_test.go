package rekor

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestServer(routes map[string]string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseFile, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}

		f, err := os.Open(responseFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	return ts
}

func TestGetLogEntry(t *testing.T) {
	ts := newTestServer(map[string]string{
		`/api/v1/log`:         `testdata/rekor-api-log.json`,
		`/api/v1/log/entries`: `testdata/rekor-api-log-entry.json`,
	})

	rc, err := NewClient(ts.URL, -1)
	if err != nil {
		t.Fatal(err)
	}

	entry, err := rc.GetEntry(1)
	if err != nil {
		t.Fatal(err)
	}

	if kind := entry.Body["kind"].(string); kind != `rekord` {
		t.Error(`expected rekord type`)
	}
	if version := entry.Body["apiVersion"].(string); version != `0.0.1` {
		t.Error(`expected api version 0.0.1`)
	}
}

func TestGetTreeState(t *testing.T) {
	ts := newTestServer(map[string]string{
		`/api/v1/log`: `testdata/rekor-api-log.json`,
	})

	rc, err := NewClient(ts.URL, -1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = rc.GetTreeState()
	if err != nil {
		t.Fatal(err)
	}
}
