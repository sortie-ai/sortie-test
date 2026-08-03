package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testStore() *Store {
	return NewStore(
		Item{ID: "1", Title: "write the spec"},
		Item{ID: "2", Title: "review the spec", Done: true},
		Item{ID: "3", Title: "ship it"},
	)
}

func TestListItems(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()

	NewRouter(testStore()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got listResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if len(got.Items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(got.Items))
	}

	if got.Items[0].ID != "1" {
		t.Errorf("items[0].id = %q, want %q", got.Items[0].ID, "1")
	}
}

func TestCreateItem(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid item",
			body:       `{"id":"4","title":"add pagination"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "malformed json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_body",
		},
		{
			name:       "missing id",
			body:       `{"title":"add pagination"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "missing_id",
		},
		{
			name:       "missing title",
			body:       `{"id":"4"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "missing_title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			NewRouter(testStore()).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantCode == "" {
				return
			}

			var got errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			if got.Error.Code != tt.wantCode {
				t.Errorf("error.code = %q, want %q", got.Error.Code, tt.wantCode)
			}
		})
	}
}
