package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"quriny.dev/internal/dsl"
)

// dslTypeToSQL maps DSL field types to PostgreSQL column types.
// This is the single source of truth for how Quriny DSL types translate into
// SQL storage. Add new DSL types here as the schema evolves.
var dslTypeToSQL = map[string]string{
	"uuid":    "UUID",
	"string":  "TEXT",
	"number":  "NUMERIC",
	"boolean": "BOOLEAN",
	"date":    "DATE",
	"text":    "TEXT", // alias for long-form string
}

// SQLTypeFor returns the PostgreSQL column type for a given DSL field type.
// Returns an error if the DSL type has no known SQL mapping.
func SQLTypeFor(dslType string) (string, error) {
	sqlType, ok := dslTypeToSQL[strings.ToLower(dslType)]
	if !ok {
		return "", fmt.Errorf("unsupported DSL field type %q: no SQL mapping defined", dslType)
	}
	return sqlType, nil
}

// ProvisionEntities creates PostgreSQL tables for all entities defined in the
// DSL model. It uses CREATE TABLE IF NOT EXISTS so it is safe to call
// repeatedly — existing tables will not be dropped or altered.
//
// Each entity becomes a table named after the entity (lowercased), and each
// field becomes a column with the mapped SQL type. The "id" field is always
// created as the PRIMARY KEY.
func (db *DB) ProvisionEntities(ctx context.Context, model *dsl.AppModel) error {
	for _, entity := range model.Entities {
		sql, err := buildCreateTableSQL(entity)
		if err != nil {
			return fmt.Errorf("build SQL for entity %s: %w", entity.Name, err)
		}

		slog.Info("provisioning table",
			slog.String("entity", entity.Name),
			slog.String("sql", sql),
		)

		if _, err := db.Pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("create table for entity %s: %w", entity.Name, err)
		}

		slog.Info("table provisioned", slog.String("entity", entity.Name))
	}

	return nil
}

// buildCreateTableSQL generates a CREATE TABLE IF NOT EXISTS statement from a
// DSL entity definition.
//
// Example output for a Product entity with fields [id:uuid, name:string, price:number]:
//
//	CREATE TABLE IF NOT EXISTS product (
//	    id UUID PRIMARY KEY,
//	    name TEXT,
//	    price NUMERIC
//	);
func buildCreateTableSQL(entity dsl.Entity) (string, error) {
	if len(entity.Fields) == 0 {
		return "", fmt.Errorf("entity %s has no fields", entity.Name)
	}

	// Use lowercase table name to follow PostgreSQL naming conventions.
	tableName := strings.ToLower(entity.Name)

	var columns []string
	for _, field := range entity.Fields {
		sqlType, err := SQLTypeFor(field.Type)
		if err != nil {
			return "", fmt.Errorf("field %s.%s: %w", entity.Name, field.Name, err)
		}

		// Column name is lowercased for consistency.
		colName := strings.ToLower(field.Name)

		// The "id" field is always the primary key.
		if colName == "id" {
			columns = append(columns, fmt.Sprintf("    %s %s PRIMARY KEY", colName, sqlType))
		} else {
			columns = append(columns, fmt.Sprintf("    %s %s", colName, sqlType))
		}
	}

	sql := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (\n%s\n);",
		tableName,
		strings.Join(columns, ",\n"),
	)

	return sql, nil
}
