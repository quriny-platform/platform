package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"quriny.dev/internal/dsl"
)

// PostgresStore is a PostgreSQL-backed implementation of EntityStore. It
// dynamically builds SQL queries based on the DSL entity definitions, so
// adding a new entity to the DSL automatically gives it full CRUD support
// without any hand-written SQL.
type PostgresStore struct {
	pool *pgxpool.Pool

	// entities maps entity name → entity definition for query building.
	entities map[string]dsl.Entity
}

// Compile-time check: PostgresStore must satisfy the EntityStore interface.
var _ EntityStore = (*PostgresStore)(nil)

// NewPostgresStore creates a PostgreSQL-backed store. It requires a live
// connection pool and the loaded DSL model so it knows which entities and
// fields exist.
func NewPostgresStore(pool *pgxpool.Pool, model *dsl.AppModel) *PostgresStore {
	entities := make(map[string]dsl.Entity, len(model.Entities))
	for _, entity := range model.Entities {
		entities[entity.Name] = entity
	}

	return &PostgresStore{
		pool:     pool,
		entities: entities,
	}
}

// List returns all records for the given entity by querying its table.
func (s *PostgresStore) List(entityName string) ([]map[string]any, error) {
	entity, err := s.resolveEntity(entityName)
	if err != nil {
		return nil, err
	}

	// Build column list from entity fields (e.g. "id, name, price").
	columns := fieldNames(entity)
	tableName := strings.ToLower(entityName)

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)

	rows, err := s.pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", entityName, err)
	}
	defer rows.Close()

	// Scan each row into a map using pgx's CollectRows with RowToMap.
	items, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return nil, fmt.Errorf("scan %s rows: %w", entityName, err)
	}

	// Convert UUID values to strings for JSON compatibility.
	for i := range items {
		normalizeRow(items[i])
	}

	return items, nil
}

// Get returns a single record identified by entity name and record ID.
func (s *PostgresStore) Get(entityName, recordID string) (map[string]any, error) {
	entity, err := s.resolveEntity(entityName)
	if err != nil {
		return nil, err
	}

	columns := fieldNames(entity)
	tableName := strings.ToLower(entityName)

	query := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", strings.Join(columns, ", "), tableName)

	rows, err := s.pool.Query(context.Background(), query, recordID)
	if err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", entityName, recordID, err)
	}
	defer rows.Close()

	record, err := pgx.CollectExactlyOneRow(rows, pgx.RowToMap)
	if err != nil {
		// pgx returns ErrNoRows when no record is found.
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("%w: %s/%s", ErrRecordNotFound, entityName, recordID)
		}
		return nil, fmt.Errorf("scan %s/%s: %w", entityName, recordID, err)
	}

	normalizeRow(record)
	return record, nil
}

// Create inserts a new record into the entity's table and returns the stored
// version. If no "id" is provided, PostgreSQL's gen_random_uuid() generates one.
func (s *PostgresStore) Create(entityName string, record map[string]any) (map[string]any, error) {
	entity, err := s.resolveEntity(entityName)
	if err != nil {
		return nil, err
	}

	tableName := strings.ToLower(entityName)

	// Build the column list and value placeholders from the provided fields.
	// We also collect the values in the same order for the parameterised query.
	var columns []string
	var placeholders []string
	var values []any
	paramIdx := 1

	for _, field := range entity.Fields {
		colName := strings.ToLower(field.Name)

		val, provided := record[field.Name]
		if !provided && colName == "id" {
			// Auto-generate a UUID if the caller didn't provide one.
			columns = append(columns, colName)
			placeholders = append(placeholders, "gen_random_uuid()")
			continue
		}

		if !provided {
			continue // Skip fields not included in the request.
		}

		columns = append(columns, colName)
		placeholders = append(placeholders, fmt.Sprintf("$%d", paramIdx))
		values = append(values, val)
		paramIdx++
	}

	// Build: INSERT INTO product (id, name, price) VALUES (gen_random_uuid(), $1, $2) RETURNING *
	returnColumns := fieldNames(entity)
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(returnColumns, ", "),
	)

	rows, err := s.pool.Query(context.Background(), query, values...)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", entityName, err)
	}
	defer rows.Close()

	created, err := pgx.CollectExactlyOneRow(rows, pgx.RowToMap)
	if err != nil {
		return nil, fmt.Errorf("scan created %s: %w", entityName, err)
	}

	normalizeRow(created)
	return created, nil
}

// Update replaces all fields of an existing record. Uses an UPDATE ... SET
// statement targeting the record by ID.
func (s *PostgresStore) Update(entityName, recordID string, record map[string]any) (map[string]any, error) {
	entity, err := s.resolveEntity(entityName)
	if err != nil {
		return nil, err
	}

	tableName := strings.ToLower(entityName)

	// Build SET clauses for each provided field (excluding "id").
	var setClauses []string
	var values []any
	paramIdx := 1

	for _, field := range entity.Fields {
		colName := strings.ToLower(field.Name)
		if colName == "id" {
			continue // Never update the primary key.
		}

		val, provided := record[field.Name]
		if !provided {
			continue
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", colName, paramIdx))
		values = append(values, val)
		paramIdx++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no updatable fields provided for %s/%s", entityName, recordID)
	}

	// The record ID is the last parameter in the WHERE clause.
	values = append(values, recordID)

	returnColumns := fieldNames(entity)
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d RETURNING %s",
		tableName,
		strings.Join(setClauses, ", "),
		paramIdx,
		strings.Join(returnColumns, ", "),
	)

	rows, err := s.pool.Query(context.Background(), query, values...)
	if err != nil {
		return nil, fmt.Errorf("update %s/%s: %w", entityName, recordID, err)
	}
	defer rows.Close()

	updated, err := pgx.CollectExactlyOneRow(rows, pgx.RowToMap)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("%w: %s/%s", ErrRecordNotFound, entityName, recordID)
		}
		return nil, fmt.Errorf("scan updated %s/%s: %w", entityName, recordID, err)
	}

	normalizeRow(updated)
	return updated, nil
}

// Delete removes a record from the entity's table. Returns ErrRecordNotFound
// if no row was affected (i.e. the record doesn't exist).
func (s *PostgresStore) Delete(entityName, recordID string) error {
	if _, err := s.resolveEntity(entityName); err != nil {
		return err
	}

	tableName := strings.ToLower(entityName)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	result, err := s.pool.Exec(context.Background(), query, recordID)
	if err != nil {
		return fmt.Errorf("delete %s/%s: %w", entityName, recordID, err)
	}

	// Check that a row was actually deleted.
	if result.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s/%s", ErrRecordNotFound, entityName, recordID)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// resolveEntity looks up the entity definition by name. Returns
// ErrEntityNotFound if the entity is not in the DSL model.
func (s *PostgresStore) resolveEntity(entityName string) (dsl.Entity, error) {
	entity, ok := s.entities[entityName]
	if !ok {
		return dsl.Entity{}, fmt.Errorf("%w: %s", ErrEntityNotFound, entityName)
	}
	return entity, nil
}

// fieldNames returns the lowercased column names for all fields in the entity.
// The order matches the DSL definition, ensuring consistent column ordering.
func fieldNames(entity dsl.Entity) []string {
	names := make([]string, len(entity.Fields))
	for i, field := range entity.Fields {
		names[i] = strings.ToLower(field.Name)
	}
	return names
}

// normalizeRow converts special PostgreSQL types (like [16]byte UUIDs) to
// their string representation for JSON serialisation compatibility.
func normalizeRow(row map[string]any) {
	for key, val := range row {
		switch v := val.(type) {
		case [16]byte:
			// pgx returns UUID columns as [16]byte; convert to string format.
			row[key] = fmt.Sprintf(
				"%08x-%04x-%04x-%04x-%012x",
				v[0:4], v[4:6], v[6:8], v[8:10], v[10:16],
			)
		}
	}
}
