package dsl

import (
	"fmt"
	"strings"
)

// validFieldTypes lists all DSL field types the platform recognises. Any field
// type not in this set will be rejected during validation, preventing typos
// (e.g. "strin" instead of "string") from silently propagating to the runtime.
var validFieldTypes = map[string]bool{
	"uuid":    true,
	"string":  true,
	"text":    true,
	"number":  true,
	"boolean": true,
	"date":    true,
}

// validActionTypes lists the CRUD action types the runtime can execute. Custom
// action types may be added in a future milestone (M7 — workflows).
var validActionTypes = map[string]bool{
	"create": true,
	"read":   true,
	"update": true,
	"delete": true,
}

// ValidateModel performs structural checks on the DSL model before it is
// transformed into the IR graph. Validation catches errors early so they
// surface at load time rather than at runtime.
//
// Checks performed:
//   - Entity names are non-empty and unique
//   - Entity fields have non-empty names with valid types
//   - Entity field names are unique within each entity
//   - Page names are non-empty and unique
//   - Page paths are non-empty and unique
//   - Component names are non-empty and unique
//   - Component entity references point to defined entities
//   - Component field references exist on the bound entity
//   - Component action references point to defined actions
//   - Action names are non-empty and unique
//   - Action types are valid CRUD types
//   - Action entity references point to defined entities
//   - Navigation page references point to defined pages
func ValidateModel(model *AppModel) error {
	// Build lookup maps for cross-reference validation.
	entityNames := map[string]bool{}
	entityFieldMap := map[string]map[string]bool{} // entity name → set of field names
	actionNames := map[string]bool{}
	pageNames := map[string]bool{}
	componentNames := map[string]bool{}

	// --- Validate entities ---
	for _, entity := range model.Entities {
		if entity.Name == "" {
			return fmt.Errorf("entity name cannot be empty")
		}

		if entityNames[entity.Name] {
			return fmt.Errorf("duplicate entity: %s", entity.Name)
		}
		entityNames[entity.Name] = true

		// Validate fields within the entity.
		fieldNames := map[string]bool{}
		for _, field := range entity.Fields {
			if field.Name == "" {
				return fmt.Errorf("entity %s: field name cannot be empty", entity.Name)
			}

			if fieldNames[field.Name] {
				return fmt.Errorf("entity %s: duplicate field %q", entity.Name, field.Name)
			}
			fieldNames[field.Name] = true

			if !validFieldTypes[strings.ToLower(field.Type)] {
				return fmt.Errorf("entity %s: field %q has unsupported type %q (allowed: uuid, string, text, number, boolean, date)",
					entity.Name, field.Name, field.Type)
			}
		}

		entityFieldMap[entity.Name] = fieldNames
	}

	// --- Validate actions ---
	for _, action := range model.Actions {
		if action.Name == "" {
			return fmt.Errorf("action name cannot be empty")
		}

		if actionNames[action.Name] {
			return fmt.Errorf("duplicate action: %s", action.Name)
		}
		actionNames[action.Name] = true

		if !validActionTypes[strings.ToLower(action.Type)] {
			return fmt.Errorf("action %s: unsupported type %q (allowed: create, read, update, delete)",
				action.Name, action.Type)
		}

		if action.Entity != "" && !entityNames[action.Entity] {
			return fmt.Errorf("action %s references unknown entity %s",
				action.Name, action.Entity)
		}
	}

	// --- Validate pages ---
	pagePaths := map[string]bool{}

	for _, page := range model.Pages {
		if page.Name == "" {
			return fmt.Errorf("page name cannot be empty")
		}

		if pageNames[page.Name] {
			return fmt.Errorf("duplicate page: %s", page.Name)
		}
		pageNames[page.Name] = true

		if page.Path == "" {
			return fmt.Errorf("page %s: path cannot be empty", page.Name)
		}

		if pagePaths[page.Path] {
			return fmt.Errorf("duplicate page path: %s", page.Path)
		}
		pagePaths[page.Path] = true

		// Validate that component references exist.
		for _, compRef := range page.Components {
			if !componentNames[compRef] {
				// Components may be defined after pages in the JSON, so we
				// defer component existence checks to after all components are
				// registered. For now, just note the reference.
			}
		}
	}

	// --- Validate components ---
	for _, component := range model.Components {
		if component.Name == "" {
			return fmt.Errorf("component name cannot be empty")
		}

		if componentNames[component.Name] {
			return fmt.Errorf("duplicate component: %s", component.Name)
		}
		componentNames[component.Name] = true

		// Component entity reference must point to a defined entity.
		if component.Entity != "" {
			if !entityNames[component.Entity] {
				return fmt.Errorf("component %s references unknown entity %s",
					component.Name, component.Entity)
			}

			// Component field references must exist on the bound entity.
			entityFields := entityFieldMap[component.Entity]
			for _, fieldRef := range component.Fields {
				if !entityFields[fieldRef] {
					return fmt.Errorf("component %s references unknown field %q on entity %s",
						component.Name, fieldRef, component.Entity)
				}
			}
		}

		// Component action references must point to defined actions.
		for _, actionRef := range component.Actions {
			if !actionNames[actionRef] {
				return fmt.Errorf("component %s references unknown action %s",
					component.Name, actionRef)
			}
		}
	}

	// --- Validate page → component references (deferred check) ---
	for _, page := range model.Pages {
		for _, compRef := range page.Components {
			if !componentNames[compRef] {
				return fmt.Errorf("page %s references unknown component %s",
					page.Name, compRef)
			}
		}
	}

	// --- Validate navigation ---
	for _, nav := range model.Navigation {
		if nav.Label == "" {
			return fmt.Errorf("navigation entry must have a label")
		}

		if nav.Page != "" && !pageNames[nav.Page] {
			return fmt.Errorf("navigation %q references unknown page %s",
				nav.Label, nav.Page)
		}
	}

	return nil
}
