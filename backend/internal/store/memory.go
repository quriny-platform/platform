package store

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"

	"quriny.dev/internal/dsl"
)

// MemoryStore is an in-memory implementation of EntityStore. It is useful for
// development, testing, and quick iteration without needing a running database.
// Data is lost when the process exits.
type MemoryStore struct {
	mu sync.RWMutex

	// entities maps entity name → entity definition for field validation.
	entities map[string]dsl.Entity

	// records maps entity name → record ID → record data.
	records map[string]map[string]map[string]any
}

// Compile-time check: MemoryStore must satisfy the EntityStore interface.
var _ EntityStore = (*MemoryStore)(nil)

// NewMemoryStore initialises a MemoryStore from a loaded DSL model. Each entity
// in the model is registered so Create/Read/Update/Delete calls can validate
// against the entity definition.
func NewMemoryStore(model *dsl.AppModel) *MemoryStore {
	s := &MemoryStore{
		entities: make(map[string]dsl.Entity, len(model.Entities)),
		records:  make(map[string]map[string]map[string]any, len(model.Entities)),
	}

	for _, entity := range model.Entities {
		s.entities[entity.Name] = entity
		s.records[entity.Name] = map[string]map[string]any{}
	}

	return s
}

// List returns every record stored under the given entity name.
func (s *MemoryStore) List(entityName string) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.entities[entityName]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrEntityNotFound, entityName)
	}

	items := make([]map[string]any, 0, len(s.records[entityName]))
	for _, record := range s.records[entityName] {
		items = append(items, cloneRecord(record))
	}

	return items, nil
}

// Get retrieves a single record by entity name and record ID.
func (s *MemoryStore) Get(entityName, recordID string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.entities[entityName]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrEntityNotFound, entityName)
	}

	record, ok := s.records[entityName][recordID]
	if !ok {
		return nil, fmt.Errorf("%w: %s/%s", ErrRecordNotFound, entityName, recordID)
	}

	return cloneRecord(record), nil
}

// Create adds a new record to the entity. If the record does not include an
// "id" field and the entity defines one, a UUID v4 is generated automatically.
func (s *MemoryStore) Create(entityName string, record map[string]any) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entity, ok := s.entities[entityName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrEntityNotFound, entityName)
	}

	normalized, recordID, err := normalizeRecord(entity, record, "")
	if err != nil {
		return nil, err
	}

	s.records[entityName][recordID] = normalized

	return cloneRecord(normalized), nil
}

// Update replaces all fields of an existing record with the given data.
func (s *MemoryStore) Update(entityName, recordID string, record map[string]any) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entity, ok := s.entities[entityName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrEntityNotFound, entityName)
	}

	if _, ok := s.records[entityName][recordID]; !ok {
		return nil, fmt.Errorf("%w: %s/%s", ErrRecordNotFound, entityName, recordID)
	}

	normalized, _, err := normalizeRecord(entity, record, recordID)
	if err != nil {
		return nil, err
	}

	s.records[entityName][recordID] = normalized

	return cloneRecord(normalized), nil
}

// Delete removes a record from the entity. Returns ErrRecordNotFound if the
// record does not exist.
func (s *MemoryStore) Delete(entityName, recordID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.entities[entityName]; !ok {
		return fmt.Errorf("%w: %s", ErrEntityNotFound, entityName)
	}

	if _, ok := s.records[entityName][recordID]; !ok {
		return fmt.Errorf("%w: %s/%s", ErrRecordNotFound, entityName, recordID)
	}

	delete(s.records[entityName], recordID)
	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// normalizeRecord validates the record against the entity definition, strips
// unknown fields, and ensures an ID is present.
func normalizeRecord(entity dsl.Entity, record map[string]any, forcedID string) (map[string]any, string, error) {
	// Build a lookup of allowed fields for O(1) validation.
	allowedFields := make(map[string]dsl.Field, len(entity.Fields))
	for _, field := range entity.Fields {
		allowedFields[field.Name] = field
	}

	// Copy only known fields into the normalised output.
	normalized := make(map[string]any, len(record)+1)
	for name, value := range record {
		field, ok := allowedFields[name]
		if !ok {
			return nil, "", fmt.Errorf("unknown field %q for entity %s", name, entity.Name)
		}

		normalized[name] = normalizeValue(field, value)
	}

	// Determine the record ID: use the forced value, or the record's "id", or generate one.
	recordID := forcedID
	if recordID == "" {
		if rawID, ok := normalized["id"]; ok {
			recordID = strings.TrimSpace(fmt.Sprint(rawID))
		}
	}

	if recordID == "" && hasField(entity, "id") {
		recordID = newRecordID()
	}

	if recordID == "" {
		return nil, "", fmt.Errorf("entity %s requires an id field", entity.Name)
	}

	// Guard against ID mismatches (e.g. body says id=X but URL says id=Y).
	if rawID, ok := normalized["id"]; ok && strings.TrimSpace(fmt.Sprint(rawID)) != recordID {
		return nil, "", fmt.Errorf("record id does not match requested id")
	}

	if hasField(entity, "id") {
		normalized["id"] = recordID
	}

	return normalized, recordID, nil
}

// normalizeValue coerces a value to the expected Go type based on the DSL field type.
func normalizeValue(field dsl.Field, value any) any {
	switch field.Type {
	case "uuid", "string":
		return strings.TrimSpace(fmt.Sprint(value))
	default:
		return value
	}
}

// hasField returns true if the entity defines a field with the given name.
func hasField(entity dsl.Entity, fieldName string) bool {
	for _, field := range entity.Fields {
		if field.Name == fieldName {
			return true
		}
	}

	return false
}

// cloneRecord creates a shallow copy of a record map so callers cannot mutate
// the store's internal data.
func cloneRecord(record map[string]any) map[string]any {
	cloned := make(map[string]any, len(record))
	for key, value := range record {
		cloned[key] = value
	}

	return cloned
}

// newRecordID generates a random UUID v4 string.
func newRecordID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		panic(err)
	}

	// Set UUID version 4 and variant bits per RFC 4122.
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		raw[0:4],
		raw[4:6],
		raw[6:8],
		raw[8:10],
		raw[10:16],
	)
}
