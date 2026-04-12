// Package api contains the HTTP-facing surface for Quriny platform and runtime
// services. It translates HTTP requests into store operations and serialises
// responses as JSON.
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"quriny.dev/internal/dsl"
	"quriny.dev/internal/ir"
	"quriny.dev/internal/store"
)

// Server exposes the current app model and runtime graph over HTTP.
// It delegates persistence to any store.EntityStore implementation,
// making it agnostic to the underlying storage engine.
type Server struct {
	model   *dsl.AppModel
	graph   *ir.AppGraph
	store   store.EntityStore
	actions *actionResolver
	mux     *http.ServeMux
}

// HealthResponse is returned by the health endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// NewServer wires the runtime HTTP API with the given model and store.
// The store parameter accepts any EntityStore implementation (in-memory,
// PostgreSQL, etc.), following the dependency inversion principle.
func NewServer(model *dsl.AppModel, graph *ir.AppGraph, entityStore store.EntityStore) *Server {
	server := &Server{
		model:   model,
		graph:   graph,
		store:   entityStore,
		actions: newActionResolver(model),
		mux:     http.NewServeMux(),
	}

	server.routes()

	return server
}

// Handler returns the configured HTTP handler for the runtime server.
// This is the entry point for http.ListenAndServe or testing via httptest.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// routes registers all HTTP endpoints on the internal mux.
func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/app-model", s.handleAppModel)
	s.mux.HandleFunc("/app-graph", s.handleAppGraph)
	s.mux.HandleFunc("/entities/", s.handleEntities)
}

// handleHealth returns a simple status check for load balancers and probes.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// handleAppModel serves the full DSL model as JSON so the runtime UI can
// discover entities, pages, components, and navigation at startup.
func (s *Server) handleAppModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, s.model)
}

// handleAppGraph serves the IR graph so clients can inspect the runtime
// topology (nodes and edges between entities, pages, components, actions).
func (s *Server) handleAppGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	writeJSON(w, http.StatusOK, s.graph)
}

// handleEntities is the dynamic CRUD handler. It parses the URL to determine
// the entity name and optional record ID, then dispatches to the store.
//
// Routes:
//   GET    /entities/{name}       → List all records
//   POST   /entities/{name}       → Create a new record
//   GET    /entities/{name}/{id}  → Get a single record
//   PUT    /entities/{name}/{id}  → Update a record
//   DELETE /entities/{name}/{id}  → Delete a record
func (s *Server) handleEntities(w http.ResponseWriter, r *http.Request) {
	// Parse URL path segments: /entities/{entityName}[/{recordID}]
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/entities/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	entityName := parts[0]

	// Resolve which HTTP methods are allowed for this entity based on DSL actions.
	collectionMethods, recordMethods := s.actions.AllowedMethodsForEntity(entityName)

	// Collection-level operations: GET (list) and POST (create).
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			// GET /entities/{name} → requires a "read" action in the DSL.
			if !s.actions.IsAllowed(entityName, actionRead) {
				writeMethodNotAllowed(w, collectionMethods...)
				return
			}

			items, err := s.store.List(entityName)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}

			writeJSON(w, http.StatusOK, map[string]any{
				"entity": entityName,
				"count":  len(items),
				"items":  items,
			})
			return

		case http.MethodPost:
			// POST /entities/{name} → requires a "create" action in the DSL.
			if !s.actions.IsAllowed(entityName, actionCreate) {
				writeMethodNotAllowed(w, collectionMethods...)
				return
			}

			record, err := decodeRecord(r)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}

			created, err := s.store.Create(entityName, record)
			if err != nil {
				writeError(w, store.StatusCodeForError(err), err.Error())
				return
			}

			writeJSON(w, http.StatusCreated, created)
			return
		}

		writeMethodNotAllowed(w, collectionMethods...)
		return
	}

	// Record-level operations: GET (read), PUT (update), DELETE (remove).
	if len(parts) == 2 {
		recordID := parts[1]

		switch r.Method {
		case http.MethodGet:
			// GET /entities/{name}/{id} → requires a "read" action in the DSL.
			if !s.actions.IsAllowed(entityName, actionRead) {
				writeMethodNotAllowed(w, recordMethods...)
				return
			}

			record, err := s.store.Get(entityName, recordID)
			if err != nil {
				writeError(w, store.StatusCodeForError(err), err.Error())
				return
			}

			writeJSON(w, http.StatusOK, record)
			return

		case http.MethodPut:
			// PUT /entities/{name}/{id} → requires an "update" action in the DSL.
			if !s.actions.IsAllowed(entityName, actionUpdate) {
				writeMethodNotAllowed(w, recordMethods...)
				return
			}

			record, err := decodeRecord(r)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}

			updated, err := s.store.Update(entityName, recordID, record)
			if err != nil {
				writeError(w, store.StatusCodeForError(err), err.Error())
				return
			}

			writeJSON(w, http.StatusOK, updated)
			return

		case http.MethodDelete:
			// DELETE /entities/{name}/{id} → requires a "delete" action in the DSL.
			if !s.actions.IsAllowed(entityName, actionDelete) {
				writeMethodNotAllowed(w, recordMethods...)
				return
			}

			if err := s.store.Delete(entityName, recordID); err != nil {
				writeError(w, store.StatusCodeForError(err), err.Error())
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		writeMethodNotAllowed(w, recordMethods...)
		return
	}

	http.NotFound(w, r)
}

// decodeRecord reads and parses a JSON object from the request body.
func decodeRecord(r *http.Request) (map[string]any, error) {
	defer r.Body.Close()

	var record map[string]any
	if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
		return nil, err
	}

	if record == nil {
		record = map[string]any{}
	}

	return record, nil
}

// writeJSON serialises value as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(append(body, '\n'))
}

// writeError writes a JSON error response with the given status code and message.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}

// writeMethodNotAllowed replies with 405 and sets the Allow header to indicate
// which HTTP methods are supported for the endpoint.
func writeMethodNotAllowed(w http.ResponseWriter, allowedMethods ...string) {
	w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
