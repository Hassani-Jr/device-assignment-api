package database

import (
	"database/sql"
	"fmt"

	"device-assignment-api/internal/config"

	_ "github.com/lib/pq"
)

// PostgresDB wraps a PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// DB returns the underlying sql.DB instance
func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// RunMigrations executes database migrations
func (p *PostgresDB) RunMigrations() error {
	migrations := []string{
		createDevicesTable,
		createAssignmentsTable,
		createIndexes,
	}

	for _, migration := range migrations {
		if _, err := p.db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

const createDevicesTable = `
CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY,
    certificate_serial_number VARCHAR(255) NOT NULL UNIQUE,
    certificate_issuer_cn VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);`

const createAssignmentsTable = `
CREATE TABLE IF NOT EXISTS assignments (
    id UUID PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    unassigned_at TIMESTAMP WITH TIME ZONE NULL
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_devices_serial_number ON devices(certificate_serial_number);
CREATE INDEX IF NOT EXISTS idx_assignments_device_id ON assignments(device_id);
CREATE INDEX IF NOT EXISTS idx_assignments_user_id ON assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_assignments_active ON assignments(device_id) WHERE unassigned_at IS NULL;`
