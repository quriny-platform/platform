package dsl

import (
	"encoding/json"
	"fmt"
)

// LoadModel parses raw DSL JSON and validates it before the runtime uses it.
func LoadModel(data []byte) (*AppModel, error) {
	var model AppModel

	err := json.Unmarshal(data, &model)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSL JSON: %w", err)
	}

	err = ValidateModel(&model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}
