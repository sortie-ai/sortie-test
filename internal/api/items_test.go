package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
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

	if got.NextCursor != nil {
		t.Errorf("next_cursor = %v, want nil", *got.NextCursor)
	}
}

func TestListItemsPagination(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		wantIDs        []string
		wantNextCursor *string
	}{
		{
			name:           "first page",
			query:          "?limit=2",
			wantIDs:        []string{"1", "2"},
			wantNextCursor: ptr("2"),
		},
		{
			name:           "second page",
			query:          "?limit=2&cursor=2",
			wantIDs:        []string{"3"},
			wantNextCursor: nil,
		},
		{
			name:           "cursor past last id",
			query:          "?cursor=3",
			wantIDs:        []string{},
			wantNextCursor: nil,
		},
		{
			name:           "default limit returns every item",
			query:          "",
			wantIDs:        []string{"1", "2", "3"},
			wantNextCursor: nil,
		},
		{
			name:           "limit exactly matching remaining items",
			query:          "?limit=3",
			wantIDs:        []string{"1", "2", "3"},
			wantNextCursor: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/items"+tt.query, nil)
			rec := httptest.NewRecorder()

			NewRouter(testStore()).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			var got listResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			gotIDs := make([]string, len(got.Items))
			for i, it := range got.Items {
				gotIDs[i] = it.ID
			}

			if !slices.Equal(gotIDs, tt.wantIDs) {
				t.Errorf("ids = %v, want %v", gotIDs, tt.wantIDs)
			}

			switch {
			case tt.wantNextCursor == nil && got.NextCursor != nil:
				t.Errorf("next_cursor = %q, want nil", *got.NextCursor)
			case tt.wantNextCursor != nil && got.NextCursor == nil:
				t.Errorf("next_cursor = nil, want %q", *tt.wantNextCursor)
			case tt.wantNextCursor != nil && *got.NextCursor != *tt.wantNextCursor:
				t.Errorf("next_cursor = %q, want %q", *got.NextCursor, *tt.wantNextCursor)
			}
		})
	}
}

func TestListItemsInvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "not a number", query: "?limit=abc"},
		{name: "zero", query: "?limit=0"},
		{name: "negative", query: "?limit=-1"},
		{name: "above maximum", query: "?limit=101"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/items"+tt.query, nil)
			rec := httptest.NewRecorder()

			NewRouter(testStore()).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			var got errorResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}

			if got.Error.Code != "invalid_limit" {
				t.Errorf("error.code = %q, want %q", got.Error.Code, "invalid_limit")
			}
		})
	}
}

func ptr(s string) *string { return &s }

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
