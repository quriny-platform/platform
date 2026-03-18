package dsl

import "fmt"

// ValidateModel performs the basic structural checks needed before the model
// can be transformed into the IR graph.
func ValidateModel(model *AppModel) error {
	entityNames := map[string]bool{}

	for _, entity := range model.Entities {
		if entity.Name == "" {
			return fmt.Errorf("entity name cannot be empty")
		}

		if entityNames[entity.Name] {
			return fmt.Errorf("duplicate entity: %s", entity.Name)
		}

		entityNames[entity.Name] = true
	}

	for _, component := range model.Components {
		if component.Entity != "" {
			if !entityNames[component.Entity] {
				return fmt.Errorf("component %s references unknown entity %s",
					component.Name,
					component.Entity)
			}
		}
	}

	return nil
}
