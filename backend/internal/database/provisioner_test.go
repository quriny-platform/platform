package database

import (
	"strings"
	"testing"

	"quriny.dev/internal/dsl"
)

// TestBuildCreateTableSQL verifies that the SQL generator produces correct
// CREATE TABLE statements from DSL entity definitions.
func TestBuildCreateTableSQL(t *testing.T) {
	tests := []struct {
		name    string
		entity  dsl.Entity
		wantSQL string // substring that must appear in the generated SQL
		wantErr bool
	}{
		{
			name: "basic entity with id, string, and number fields",
			entity: dsl.Entity{
				Name: "Product",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "name", Type: "string"},
					{Name: "price", Type: "number"},
				},
			},
			wantSQL: "CREATE TABLE IF NOT EXISTS product",
			wantErr: false,
		},
		{
			name: "id field gets PRIMARY KEY constraint",
			entity: dsl.Entity{
				Name: "User",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "email", Type: "string"},
				},
			},
			wantSQL: "id UUID PRIMARY KEY",
			wantErr: false,
		},
		{
			name: "boolean and date types are mapped correctly",
			entity: dsl.Entity{
				Name: "Task",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "done", Type: "boolean"},
					{Name: "dueDate", Type: "date"},
				},
			},
			wantSQL: "done BOOLEAN",
			wantErr: false,
		},
		{
			name:    "entity with no fields returns error",
			entity:  dsl.Entity{Name: "Empty", Fields: []dsl.Field{}},
			wantErr: true,
		},
		{
			name: "unsupported DSL type returns error",
			entity: dsl.Entity{
				Name: "Bad",
				Fields: []dsl.Field{
					{Name: "id", Type: "uuid"},
					{Name: "data", Type: "binary"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := buildCreateTableSQL(tt.entity)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got SQL: %s", sql)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(sql, tt.wantSQL) {
				t.Errorf("expected SQL to contain %q, got:\n%s", tt.wantSQL, sql)
			}
		})
	}
}

// TestSQLTypeFor verifies the DSL type → SQL type mapping.
func TestSQLTypeFor(t *testing.T) {
	tests := []struct {
		dslType string
		want    string
		wantErr bool
	}{
		{"uuid", "UUID", false},
		{"string", "TEXT", false},
		{"number", "NUMERIC", false},
		{"boolean", "BOOLEAN", false},
		{"date", "DATE", false},
		{"text", "TEXT", false},
		{"binary", "", true},   // unsupported type
		{"unknown", "", true},  // unsupported type
	}

	for _, tt := range tests {
		t.Run(tt.dslType, func(t *testing.T) {
			got, err := SQLTypeFor(tt.dslType)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for type %q, got %q", tt.dslType, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for type %q: %v", tt.dslType, err)
			}

			if got != tt.want {
				t.Errorf("SQLTypeFor(%q) = %q, want %q", tt.dslType, got, tt.want)
			}
		})
	}
}
