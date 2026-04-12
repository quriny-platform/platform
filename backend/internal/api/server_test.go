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

// TestServerEntityCRUDLifecycle verifies the full create → list → get → update
// → delete → verify-gone cycle through the HTTP API using an in-memory store.
func TestServerEntityCRUDLifecycle(t *testing.T) {
	model := &dsl.AppModel{
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
	}

	graph, err := ir.BuildAppGraph(model)
	if err != nil {
		t.Fatalf("build app graph: %v", err)
	}

	// Use the in-memory store implementation behind the EntityStore interface.
	memStore := store.NewMemoryStore(model)
	server := NewServer(model, graph, memStore)

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
