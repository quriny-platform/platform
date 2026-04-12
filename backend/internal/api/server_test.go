package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"quriny.dev/internal/dsl"
	"quriny.dev/internal/ir"
	"quriny.dev/internal/store"
)

// newTestServer is a helper that creates a Server with the given model and an
// in-memory store, ready for HTTP testing.
func newTestServer(t *testing.T, model *dsl.AppModel) *Server {
	t.Helper()

	graph, err := ir.BuildAppGraph(model)
	if err != nil {
		t.Fatalf("build app graph: %v", err)
	}

	memStore := store.NewMemoryStore(model)
	return NewServer(model, graph, memStore)
}

// fullCRUDModel returns a DSL model with a Product entity and all four CRUD
// actions defined. This represents the "everything allowed" baseline.
func fullCRUDModel() *dsl.AppModel {
	return &dsl.AppModel{
		Entities: []dsl.Entity{
			{
				Name: "Product",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
					{Name: "price", Type: "number"},
				},
			},
		},
		Actions: []dsl.Action{
			{Name: "CreateProduct", Type: "create", Entity: "Product"},
			{Name: "ListProducts", Type: "read", Entity: "Product"},
			{Name: "UpdateProduct", Type: "update", Entity: "Product"},
			{Name: "DeleteProduct", Type: "delete", Entity: "Product"},
		},
	}
}

// TestServerEntityCRUDLifecycle verifies the full create → list → get → update
// → delete → verify-gone cycle through the HTTP API using an in-memory store
// with all CRUD actions enabled.
func TestServerEntityCRUDLifecycle(t *testing.T) {
	server := newTestServer(t, fullCRUDModel())

	// Step 1: Create a new Product record.
	created := performRequest(
		t,
		server.Handler(),
		http.MethodPost,
		"/entities/Product",
		map[string]any{"name": "Notebook", "price": 12.5},
		http.StatusCreated,
	)

	recordID, ok := created["id"].(string)
	if !ok || recordID == "" {
		t.Fatalf("expected generated id, got %#v", created["id"])
	}

	if created["name"] != "Notebook" {
		t.Fatalf("expected created name Notebook, got %#v", created["name"])
	}

	// Step 2: List all products — should contain exactly one.
	list := performRequest(
		t,
		server.Handler(),
		http.MethodGet,
		"/entities/Product",
		nil,
		http.StatusOK,
	)

	items, ok := list["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected exactly one record, got %#v", list["items"])
	}

	// Step 3: Get the single product by ID.
	fetched := performRequest(
		t,
		server.Handler(),
		http.MethodGet,
		"/entities/Product/"+recordID,
		nil,
		http.StatusOK,
	)

	if fetched["id"] != recordID {
		t.Fatalf("expected fetched id %s, got %#v", recordID, fetched["id"])
	}

	// Step 4: Update the product.
	updated := performRequest(
		t,
		server.Handler(),
		http.MethodPut,
		"/entities/Product/"+recordID,
		map[string]any{"name": "Desk", "price": 20.0},
		http.StatusOK,
	)

	if updated["name"] != "Desk" {
		t.Fatalf("expected updated name Desk, got %#v", updated["name"])
	}

	// Step 5: Delete the product.
	request := httptest.NewRequest(http.MethodDelete, "/entities/Product/"+recordID, nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected delete status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	// Step 6: Confirm the product no longer exists.
	request = httptest.NewRequest(http.MethodGet, "/entities/Product/"+recordID, nil)
	recorder = httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected deleted record status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

// TestActionEnforcement_ReadOnlyEntity verifies that when only a "read" action
// is defined, POST/PUT/DELETE are rejected with 405 Method Not Allowed.
func TestActionEnforcement_ReadOnlyEntity(t *testing.T) {
	model := &dsl.AppModel{
		Entities: []dsl.Entity{
			{
				Name: "Product",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
				},
			},
		},
		Actions: []dsl.Action{
			// Only read — no create, update, or delete.
			{Name: "ListProducts", Type: "read", Entity: "Product"},
		},
	}

	server := newTestServer(t, model)

	// GET (list) should work.
	performRequest(t, server.Handler(), http.MethodGet, "/entities/Product", nil, http.StatusOK)

	// POST (create) should be rejected.
	assertStatus(t, server.Handler(), http.MethodPost, "/entities/Product",
		map[string]any{"name": "Blocked"}, http.StatusMethodNotAllowed)

	// PUT (update) should be rejected.
	assertStatus(t, server.Handler(), http.MethodPut, "/entities/Product/some-id",
		map[string]any{"name": "Blocked"}, http.StatusMethodNotAllowed)

	// DELETE should be rejected.
	assertStatus(t, server.Handler(), http.MethodDelete, "/entities/Product/some-id",
		nil, http.StatusMethodNotAllowed)
}

// TestActionEnforcement_CreateAndReadOnly verifies that when only "create" and
// "read" actions are defined, update and delete are blocked.
func TestActionEnforcement_CreateAndReadOnly(t *testing.T) {
	model := &dsl.AppModel{
		Entities: []dsl.Entity{
			{
				Name: "Product",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
				},
			},
		},
		Actions: []dsl.Action{
			{Name: "CreateProduct", Type: "create", Entity: "Product"},
			{Name: "ListProducts", Type: "read", Entity: "Product"},
		},
	}

	server := newTestServer(t, model)

	// Create should succeed.
	created := performRequest(t, server.Handler(), http.MethodPost, "/entities/Product",
		map[string]any{"name": "Allowed"}, http.StatusCreated)

	recordID := created["id"].(string)

	// Read should succeed.
	performRequest(t, server.Handler(), http.MethodGet, "/entities/Product/"+recordID, nil, http.StatusOK)

	// Update should be blocked (no "update" action).
	assertStatus(t, server.Handler(), http.MethodPut, "/entities/Product/"+recordID,
		map[string]any{"name": "Blocked"}, http.StatusMethodNotAllowed)

	// Delete should be blocked (no "delete" action).
	assertStatus(t, server.Handler(), http.MethodDelete, "/entities/Product/"+recordID,
		nil, http.StatusMethodNotAllowed)
}

// TestActionEnforcement_NoActions verifies that when no actions are defined
// for an entity, all CRUD operations are rejected.
func TestActionEnforcement_NoActions(t *testing.T) {
	model := &dsl.AppModel{
		Entities: []dsl.Entity{
			{
				Name: "Product",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
				},
			},
		},
		Actions: []dsl.Action{}, // no actions at all
	}

	server := newTestServer(t, model)

	assertStatus(t, server.Handler(), http.MethodGet, "/entities/Product", nil, http.StatusMethodNotAllowed)
	assertStatus(t, server.Handler(), http.MethodPost, "/entities/Product",
		map[string]any{"name": "Blocked"}, http.StatusMethodNotAllowed)
	assertStatus(t, server.Handler(), http.MethodGet, "/entities/Product/some-id", nil, http.StatusMethodNotAllowed)
	assertStatus(t, server.Handler(), http.MethodPut, "/entities/Product/some-id",
		map[string]any{"name": "Blocked"}, http.StatusMethodNotAllowed)
	assertStatus(t, server.Handler(), http.MethodDelete, "/entities/Product/some-id", nil, http.StatusMethodNotAllowed)
}

// TestActionEnforcement_AllowHeaderPresent verifies that 405 responses include
// the Allow header listing the permitted methods.
func TestActionEnforcement_AllowHeaderPresent(t *testing.T) {
	model := &dsl.AppModel{
		Entities: []dsl.Entity{
			{
				Name: "Product",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
				},
			},
		},
		Actions: []dsl.Action{
			{Name: "ListProducts", Type: "read", Entity: "Product"},
		},
	}

	server := newTestServer(t, model)

	// POST should be rejected; the Allow header should include GET (from "read").
	request := httptest.NewRequest(http.MethodPost, "/entities/Product", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", recorder.Code)
	}

	allow := recorder.Header().Get("Allow")
	if allow == "" {
		t.Fatal("expected Allow header to be set on 405 response")
	}

	if allow != "GET" {
		t.Fatalf("expected Allow header to be 'GET', got %q", allow)
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// assertStatus sends an HTTP request and asserts the response status code,
// without decoding the body.
func assertStatus(t *testing.T, handler http.Handler, method, path string, body any, wantStatus int) {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	request := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != wantStatus {
		t.Fatalf("%s %s: expected status %d, got %d (body: %s)",
			method, path, wantStatus, recorder.Code, recorder.Body.String())
	}
}

// performRequest is a test helper that sends an HTTP request and decodes the
// JSON response, failing the test on unexpected status codes.
func performRequest(t *testing.T, handler http.Handler, method, path string, body any, wantStatus int) map[string]any {
	t.Helper()

	var requestBody *bytes.Reader
	if body == nil {
		requestBody = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		requestBody = bytes.NewReader(payload)
	}

	request := httptest.NewRequest(method, path, requestBody)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != wantStatus {
		t.Fatalf("expected status %d, got %d with body %s", wantStatus, recorder.Code, recorder.Body.String())
	}

	var response map[string]any
	if recorder.Body.Len() == 0 {
		return response
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return response
}
