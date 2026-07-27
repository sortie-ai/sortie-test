package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
	if got := rec.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", got, "text/plain; charset=utf-8")
	}
}

func TestHealthzAcceptsAnyMethod(t *testing.T) {
	t.Parallel()

	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			newMux().ServeHTTP(rec, httptest.NewRequest(method, "/healthz", nil))

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if body := rec.Body.String(); body != "ok" {
				t.Errorf("body = %q, want %q", body, "ok")
			}
		})
	}
}

// TestHealthzMatchesExactPathOnly guards against registering the route as a
// subtree, which would answer every path below /healthz.
func TestHealthzMatchesExactPathOnly(t *testing.T) {
	t.Parallel()

	paths := []string{"/", "/healthz/", "/healthz/extra", "/healthzz", "/Healthz"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			newMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
			if body := rec.Body.String(); body == "ok" {
				t.Errorf("body = %q, want the route not to match", body)
			}
		})
	}
}

// TestHealthzOverHTTPServer exercises the route through a real server so the
// wiring in newMux is covered end to end, not just the handler in isolation.
func TestHealthzOverHTTPServer(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(newMux())
	// Cleanup, not defer: deferred calls would close the server before the
	// parallel subtests below resume.
	t.Cleanup(server.Close)

	t.Run("GET returns ok", func(t *testing.T) {
		t.Parallel()

		resp, err := server.Client().Get(server.URL + "/healthz")
		if err != nil {
			t.Fatalf("GET /healthz: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if string(body) != "ok" {
			t.Errorf("body = %q, want %q", body, "ok")
		}
	})

	t.Run("HEAD omits the body", func(t *testing.T) {
		t.Parallel()

		resp, err := server.Client().Head(server.URL + "/healthz")
		if err != nil {
			t.Fatalf("HEAD /healthz: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if len(body) != 0 {
			t.Errorf("body = %q, want empty", body)
		}
	})
}
