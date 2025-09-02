package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int64
	Name    string
	Up      string
	Down    string
}

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Up(migrationsFS fs.FS) error {
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := m.loadMigrations(migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`
	_, err := m.db.Exec(query)
	return err
}

func (m *Migrator) loadMigrations(migrationsFS fs.FS) ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		version, err := m.extractVersion(path)
		if err != nil {
			return err
		}

		content, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return err
		}

		migration := Migration{
			Version: version,
			Name:    filepath.Base(path),
			Up:      string(content),
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) extractVersion(filename string) (int64, error) {
	base := filepath.Base(filename)
	parts := strings.Split(base, "_")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid migration filename: %s", filename)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid version in filename %s: %w", filename, err)
	}

	return version, nil
}

func (m *Migrator) getAppliedMigrations() (map[int64]bool, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func (m *Migrator) applyMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(migration.Up); err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
		return err
	}

	return tx.Commit()
}
