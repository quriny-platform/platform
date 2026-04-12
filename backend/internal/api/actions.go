package api

import "quriny.dev/internal/dsl"

// actionType constants mirror the DSL action types. These are the only values
// recognised in the "type" field of a DSL action definition.
const (
	actionCreate = "create"
	actionRead   = "read"
	actionUpdate = "update"
	actionDelete = "delete"
)

// actionResolver determines which CRUD operations are permitted for each entity
// based on the DSL actions section. If an entity has no "delete" action defined,
// the runtime will reject DELETE requests for that entity.
//
// This enforces the principle that the DSL is the single source of truth:
// only what's declared is allowed.
type actionResolver struct {
	// allowed maps entity name → set of permitted action types.
	// Example: {"Product": {"create": true, "read": true}}
	allowed map[string]map[string]bool
}

// newActionResolver builds the resolver from the loaded DSL model. It indexes
// every action by its target entity and type for O(1) lookup at request time.
func newActionResolver(model *dsl.AppModel) *actionResolver {
	allowed := make(map[string]map[string]bool)

	for _, action := range model.Actions {
		if action.Entity == "" || action.Type == "" {
			continue // skip malformed actions
		}

		if _, ok := allowed[action.Entity]; !ok {
			allowed[action.Entity] = make(map[string]bool)
		}

		allowed[action.Entity][action.Type] = true
	}

	return &actionResolver{allowed: allowed}
}

// IsAllowed returns true if the given action type is permitted for the entity.
// For example, IsAllowed("Product", "delete") returns true only if the DSL
// declares a delete action targeting the Product entity.
func (r *actionResolver) IsAllowed(entityName, actionType string) bool {
	entityActions, ok := r.allowed[entityName]
	if !ok {
		return false // no actions defined for this entity at all
	}

	return entityActions[actionType]
}

// AllowedMethodsForEntity returns the list of HTTP methods permitted for
// collection-level and record-level endpoints. This is used to populate the
// Allow header in 405 responses so clients know what's available.
func (r *actionResolver) AllowedMethodsForEntity(entityName string) (collection []string, record []string) {
	// Collection-level: GET (list) maps to "read", POST maps to "create".
	if r.IsAllowed(entityName, actionRead) {
		collection = append(collection, "GET")
		record = append(record, "GET")
	}
	if r.IsAllowed(entityName, actionCreate) {
		collection = append(collection, "POST")
	}

	// Record-level: PUT maps to "update", DELETE maps to "delete".
	if r.IsAllowed(entityName, actionUpdate) {
		record = append(record, "PUT")
	}
	if r.IsAllowed(entityName, actionDelete) {
		record = append(record, "DELETE")
	}

	return collection, record
}
