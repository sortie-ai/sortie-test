package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// defaultItemsLimit and maxItemsLimit bound the limit query parameter on
// GET /items.
const (
	defaultItemsLimit = 20
	maxItemsLimit     = 100
)

// Item is a single task record.
type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Store keeps items in memory and is safe for concurrent use.
type Store struct {
	mu    sync.RWMutex
	items map[string]Item
}

// NewStore returns a Store seeded with the given items.
func NewStore(seed ...Item) *Store {
	s := &Store{items: make(map[string]Item, len(seed))}
	for _, it := range seed {
		s.items[it.ID] = it
	}

	return s
}

// List returns every stored item ordered by ID.
func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Item, 0, len(s.items))
	for _, it := range s.items {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })

	return out
}

// Put stores it, replacing any existing item with the same ID.
func (s *Store) Put(it Item) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[it.ID] = it
}

// listResponse wraps a collection so the payload stays an object.
type listResponse struct {
	Items      []Item  `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

// Handler serves the item endpoints.
type Handler struct {
	store *Store
}

// NewHandler returns a Handler backed by store.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// NewRouter registers every route the service exposes.
func NewRouter(store *Store) *http.ServeMux {
	h := NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", h.ListItems)
	mux.HandleFunc("POST /items", h.CreateItem)

	return mux
}

// ListItems responds with a page of stored items ordered by ID.
//
// The limit query parameter caps the page size (default 20, maximum 100).
// The cursor query parameter, when set, is the ID of the last item from the
// previous page; items with a greater ID are returned.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	limit := defaultItemsLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > maxItemsLimit {
			WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be a number between 1 and 100")

			return
		}
		limit = parsed
	}

	items := h.store.List()

	cursor := r.URL.Query().Get("cursor")
	start := 0
	if cursor != "" {
		start = len(items)
		for i, it := range items {
			if it.ID > cursor {
				start = i

				break
			}
		}
	}

	page := items[start:]

	var nextCursor *string
	if len(page) > limit {
		page = page[:limit]
		lastID := page[len(page)-1].ID
		nextCursor = &lastID
	}

	writeJSON(w, http.StatusOK, listResponse{Items: page, NextCursor: nextCursor})
}

// CreateItem stores a new item read from the request body.
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var it Item
	if err := json.NewDecoder(r.Body).Decode(&it); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "request body is not valid JSON")

		return
	}

	it.ID = strings.TrimSpace(it.ID)
	it.Title = strings.TrimSpace(it.Title)

	if it.ID == "" {
		WriteError(w, http.StatusBadRequest, "missing_id", "id is required")

		return
	}

	if it.Title == "" {
		WriteError(w, http.StatusBadRequest, "missing_title", "title is required")

		return
	}

	h.store.Put(it)
	writeJSON(w, http.StatusCreated, it)
}
