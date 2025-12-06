package db

// Schema-level constants for the internal DB structures.
// Block 4 defines only the infrastructure table used to track migrations.
// Feature-specific tables will be added in their respective blocks.

const (
	// SchemaMigrationsTable is the table that tracks applied migrations.
	SchemaMigrationsTable = "schema_migrations"
)
