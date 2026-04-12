package dsl

import (
	"strings"
	"testing"
)

// validModel returns a minimal but valid AppModel that passes all validation
// checks. Tests modify specific parts to trigger individual validation errors.
func validModel() *AppModel {
	return &AppModel{
		Entities: []Entity{
			{
				Name: "Product",
				Fields: []Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
					{Name: "price", Type: "number"},
				},
			},
		},
		Actions: []Action{
			{Name: "CreateProduct", Type: "create", Entity: "Product"},
			{Name: "ListProducts", Type: "read", Entity: "Product"},
		},
		Pages: []Page{
			{Name: "ProductList", Path: "/products", Components: []string{"ProductTable"}},
		},
		Components: []Component{
			{Name: "ProductTable", Type: "table", Entity: "Product", Fields: []string{"name", "price"}, Actions: []string{"CreateProduct"}},
		},
		Navigation: []Navigation{
			{Label: "Products", Path: "/", Page: "ProductList"},
		},
	}
}

// TestValidateModel_ValidModel confirms that a well-formed model passes validation.
func TestValidateModel_ValidModel(t *testing.T) {
	if err := ValidateModel(validModel()); err != nil {
		t.Fatalf("expected valid model to pass, got: %v", err)
	}
}

// TestValidateModel_EntityErrors tests all entity-level validation rules.
func TestValidateModel_EntityErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *AppModel)
		wantErr string
	}{
		{
			name: "empty entity name",
			mutate: func(m *AppModel) {
				m.Entities[0].Name = ""
			},
			wantErr: "entity name cannot be empty",
		},
		{
			name: "duplicate entity name",
			mutate: func(m *AppModel) {
				m.Entities = append(m.Entities, Entity{Name: "Product", Fields: []Field{{Name: "id", Type: "uuid"}}})
			},
			wantErr: "duplicate entity: Product",
		},
		{
			name: "empty field name",
			mutate: func(m *AppModel) {
				m.Entities[0].Fields[1].Name = ""
			},
			wantErr: "field name cannot be empty",
		},
		{
			name: "duplicate field name",
			mutate: func(m *AppModel) {
				m.Entities[0].Fields = append(m.Entities[0].Fields, Field{Name: "name", Type: "string"})
			},
			wantErr: "duplicate field",
		},
		{
			name: "unsupported field type",
			mutate: func(m *AppModel) {
				m.Entities[0].Fields[1].Type = "binary"
			},
			wantErr: "unsupported type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validModel()
			tt.mutate(m)
			err := ValidateModel(m)
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

// TestValidateModel_ActionErrors tests all action-level validation rules.
func TestValidateModel_ActionErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *AppModel)
		wantErr string
	}{
		{
			name: "empty action name",
			mutate: func(m *AppModel) {
				m.Actions[0].Name = ""
			},
			wantErr: "action name cannot be empty",
		},
		{
			name: "duplicate action name",
			mutate: func(m *AppModel) {
				m.Actions = append(m.Actions, Action{Name: "CreateProduct", Type: "create", Entity: "Product"})
			},
			wantErr: "duplicate action: CreateProduct",
		},
		{
			name: "unsupported action type",
			mutate: func(m *AppModel) {
				m.Actions[0].Type = "archive"
			},
			wantErr: "unsupported type",
		},
		{
			name: "action references unknown entity",
			mutate: func(m *AppModel) {
				m.Actions[0].Entity = "Order"
			},
			wantErr: "unknown entity Order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validModel()
			tt.mutate(m)
			err := ValidateModel(m)
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

// TestValidateModel_PageErrors tests all page-level validation rules.
func TestValidateModel_PageErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *AppModel)
		wantErr string
	}{
		{
			name: "empty page name",
			mutate: func(m *AppModel) {
				m.Pages[0].Name = ""
			},
			wantErr: "page name cannot be empty",
		},
		{
			name: "duplicate page name",
			mutate: func(m *AppModel) {
				m.Pages = append(m.Pages, Page{Name: "ProductList", Path: "/other", Components: []string{}})
			},
			wantErr: "duplicate page: ProductList",
		},
		{
			name: "empty page path",
			mutate: func(m *AppModel) {
				m.Pages[0].Path = ""
			},
			wantErr: "path cannot be empty",
		},
		{
			name: "duplicate page path",
			mutate: func(m *AppModel) {
				m.Pages = append(m.Pages, Page{Name: "ProductForm", Path: "/products", Components: []string{}})
			},
			wantErr: "duplicate page path",
		},
		{
			name: "page references unknown component",
			mutate: func(m *AppModel) {
				m.Pages[0].Components = []string{"NonExistent"}
			},
			wantErr: "unknown component NonExistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validModel()
			tt.mutate(m)
			err := ValidateModel(m)
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

// TestValidateModel_ComponentErrors tests all component-level validation rules.
func TestValidateModel_ComponentErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *AppModel)
		wantErr string
	}{
		{
			name: "empty component name",
			mutate: func(m *AppModel) {
				m.Components[0].Name = ""
			},
			wantErr: "component name cannot be empty",
		},
		{
			name: "duplicate component name",
			mutate: func(m *AppModel) {
				m.Components = append(m.Components, Component{Name: "ProductTable", Type: "form"})
			},
			wantErr: "duplicate component: ProductTable",
		},
		{
			name: "component references unknown entity",
			mutate: func(m *AppModel) {
				m.Components[0].Entity = "Order"
			},
			wantErr: "unknown entity Order",
		},
		{
			name: "component references unknown field on entity",
			mutate: func(m *AppModel) {
				m.Components[0].Fields = []string{"name", "description"}
			},
			wantErr: "unknown field \"description\" on entity Product",
		},
		{
			name: "component references unknown action",
			mutate: func(m *AppModel) {
				m.Components[0].Actions = []string{"ArchiveProduct"}
			},
			wantErr: "unknown action ArchiveProduct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validModel()
			tt.mutate(m)
			err := ValidateModel(m)
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

// TestValidateModel_NavigationErrors tests navigation validation rules.
func TestValidateModel_NavigationErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(m *AppModel)
		wantErr string
	}{
		{
			name: "empty navigation label",
			mutate: func(m *AppModel) {
				m.Navigation[0].Label = ""
			},
			wantErr: "must have a label",
		},
		{
			name: "navigation references unknown page",
			mutate: func(m *AppModel) {
				m.Navigation[0].Page = "Dashboard"
			},
			wantErr: "unknown page Dashboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validModel()
			tt.mutate(m)
			err := ValidateModel(m)
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

// assertErrorContains fails the test if err is nil or doesn't contain the
// expected substring.
func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}

	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got: %v", want, err)
	}
}
