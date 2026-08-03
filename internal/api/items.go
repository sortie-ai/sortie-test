package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
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
	Items []Item `json:"items"`
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

// ListItems responds with every stored item.
func (h *Handler) ListItems(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, listResponse{Items: h.store.List()})
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
