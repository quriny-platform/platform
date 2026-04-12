// Package store defines the EntityStore interface that any persistence backend
// must implement. This decouples the HTTP/API layer from specific storage
// engines (in-memory, PostgreSQL, etc.) following the dependency inversion
// principle: high-level modules depend on abstractions, not concretions.
package store

import "errors"

// Sentinel errors returned by all EntityStore implementations. Callers can use
// errors.Is() to match these regardless of which backend is in use.
var (
	ErrEntityNotFound = errors.New("entity not found")
	ErrRecordNotFound = errors.New("record not found")
)

// EntityStore is the contract every persistence backend must satisfy.
// Each method operates on records belonging to a named DSL entity.
type EntityStore interface {
	// List returns all records for the given entity.
	List(entityName string) ([]map[string]any, error)

	// Get returns a single record identified by entity name and record ID.
	Get(entityName, recordID string) (map[string]any, error)

	// Create inserts a new record and returns the stored version (with generated ID if applicable).
	Create(entityName string, record map[string]any) (map[string]any, error)

	// Update replaces the record identified by entity name and record ID with the given data.
	Update(entityName, recordID string, record map[string]any) (map[string]any, error)

	// Delete removes the record identified by entity name and record ID.
	Delete(entityName, recordID string) error
}

// StatusCodeForError maps sentinel store errors to appropriate HTTP status codes.
// This helper keeps HTTP-awareness out of the store implementations themselves.
func StatusCodeForError(err error) int {
	switch {
	case errors.Is(err, ErrEntityNotFound), errors.Is(err, ErrRecordNotFound):
		return 404
	default:
		return 400
	}
}
