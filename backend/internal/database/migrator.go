package database

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    VARCHAR(255) PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

var (
	migrationFileRegex = regexp.MustCompile(`^(\d{14}_[a-zA-Z0-9_]+)\.(up|down)\.sql$`)
	nonAlnumRegex      = regexp.MustCompile(`[^a-z0-9]+`)
)

type Migration struct {
	Version  string
	UpFile   string
	DownFile string
}

type MigrationStatus struct {
	Version   string
	Applied   bool
	AppliedAt *time.Time
}

type Migrator struct {
	pool *pgxpool.Pool
	fsys fs.FS
}

func NewMigrator(pool *pgxpool.Pool, fsys fs.FS) *Migrator {
	return &Migrator{
		pool: pool,
		fsys: fsys,
	}
}

func (m *Migrator) ensureTable(ctx context.Context) error {
	if _, err := m.pool.Exec(ctx, createMigrationsTableSQL); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) appliedVersions(ctx context.Context) (map[string]time.Time, error) {
	rows, err := m.pool.Query(ctx, `SELECT version, applied_at FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, fmt.Errorf("scan migration version: %w", err)
		}
		applied[version] = appliedAt
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return applied, nil
}

func LoadMigrations(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	byVersion := make(map[string]*Migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFileRegex.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		version, direction := matches[1], matches[2]
		mig, exists := byVersion[version]
		if !exists {
			mig = &Migration{Version: version}
			byVersion[version] = mig
		}
		if direction == "up" {
			mig.UpFile = entry.Name()
		} else {
			mig.DownFile = entry.Name()
		}
	}

	list := make([]Migration, 0, len(byVersion))
	for _, mig := range byVersion {
		list = append(list, *mig)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Version < list[j].Version
	})
	return list, nil
}

func (m *Migrator) Up(ctx context.Context) ([]string, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	all, err := LoadMigrations(m.fsys)
	if err != nil {
		return nil, err
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	executed := make([]string, 0)
	for _, mig := range all {
		if _, ok := applied[mig.Version]; ok {
			continue
		}
		if mig.UpFile == "" {
			return executed, fmt.Errorf("missing up migration file for version %s", mig.Version)
		}

		sqlBytes, err := fs.ReadFile(m.fsys, mig.UpFile)
		if err != nil {
			return executed, fmt.Errorf("read %s: %w", mig.UpFile, err)
		}

		tx, err := m.pool.Begin(ctx)
		if err != nil {
			return executed, fmt.Errorf("begin transaction for %s: %w", mig.Version, err)
		}

		if strings.TrimSpace(string(sqlBytes)) != "" {
			if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
				_ = tx.Rollback(ctx)
				return executed, fmt.Errorf("apply migration %s: %w", mig.Version, err)
			}
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, mig.Version); err != nil {
			_ = tx.Rollback(ctx)
			return executed, fmt.Errorf("record migration %s: %w", mig.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return executed, fmt.Errorf("commit migration %s: %w", mig.Version, err)
		}

		executed = append(executed, mig.Version)
	}

	return executed, nil
}

func (m *Migrator) Down(ctx context.Context, steps int) ([]string, error) {
	if steps <= 0 {
		steps = 1
	}

	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	all, err := LoadMigrations(m.fsys)
	if err != nil {
		return nil, err
	}

	byVersion := make(map[string]Migration, len(all))
	for _, mig := range all {
		byVersion[mig.Version] = mig
	}

	rows, err := m.pool.Query(
		ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT $1`,
		steps,
	)
	if err != nil {
		return nil, fmt.Errorf("query migrations to rollback: %w", err)
	}
	defer rows.Close()

	targets := make([]string, 0, steps)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan rollback version: %w", err)
		}
		targets = append(targets, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rollback versions: %w", err)
	}

	rolledBack := make([]string, 0, len(targets))
	for _, version := range targets {
		mig, ok := byVersion[version]
		if !ok || mig.DownFile == "" {
			return rolledBack, fmt.Errorf("missing down migration file for version %s", version)
		}

		sqlBytes, err := fs.ReadFile(m.fsys, mig.DownFile)
		if err != nil {
			return rolledBack, fmt.Errorf("read %s: %w", mig.DownFile, err)
		}

		tx, err := m.pool.Begin(ctx)
		if err != nil {
			return rolledBack, fmt.Errorf("begin rollback transaction for %s: %w", version, err)
		}

		if strings.TrimSpace(string(sqlBytes)) != "" {
			if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
				_ = tx.Rollback(ctx)
				return rolledBack, fmt.Errorf("rollback migration %s: %w", version, err)
			}
		}

		if _, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
			_ = tx.Rollback(ctx)
			return rolledBack, fmt.Errorf("remove migration record %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return rolledBack, fmt.Errorf("commit rollback %s: %w", version, err)
		}

		rolledBack = append(rolledBack, version)
	}

	return rolledBack, nil
}

func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	all, err := LoadMigrations(m.fsys)
	if err != nil {
		return nil, err
	}

	applied, err := m.appliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	statuses := make([]MigrationStatus, 0, len(all))
	for _, mig := range all {
		status := MigrationStatus{Version: mig.Version}
		if t, ok := applied[mig.Version]; ok {
			ts := t
			status.Applied = true
			status.AppliedAt = &ts
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

// CreateMigrationFiles generates timestamp-prefixed .up.sql and .down.sql files.
func CreateMigrationFiles(dir, name string, now time.Time) (string, string, error) {
	clean := strings.Trim(nonAlnumRegex.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "_"), "_")
	if clean == "" {
		return "", "", fmt.Errorf("migration name is required (e.g. create_users_table)")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create migrations directory: %w", err)
	}

	timestamp := now.UTC().Format("20060102150405")
	base := fmt.Sprintf("%s_%s", timestamp, clean)
	upPath := filepath.Join(dir, base+".up.sql")
	downPath := filepath.Join(dir, base+".down.sql")

	upTemplate := fmt.Sprintf("-- Migration: %s (up)\n\n", base)
	downTemplate := fmt.Sprintf("-- Migration: %s (down)\n\n", base)

	if err := os.WriteFile(upPath, []byte(upTemplate), 0o644); err != nil {
		return "", "", fmt.Errorf("write up migration: %w", err)
	}
	if err := os.WriteFile(downPath, []byte(downTemplate), 0o644); err != nil {
		return "", "", fmt.Errorf("write down migration: %w", err)
	}

	return upPath, downPath, nil
}
