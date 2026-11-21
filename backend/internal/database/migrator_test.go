package database_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"monorepo/backend/internal/database"
	"monorepo/backend/migrations"
)

func TestLoadMigrations_EmbeddedTimestampFiles(t *testing.T) {
	list, err := database.LoadMigrations(migrations.Files)
	if err != nil {
		t.Fatalf("unexpected error loading embedded migrations: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expected at least one timestamp migration")
	}
	if list[0].UpFile == "" || list[0].DownFile == "" {
		t.Fatalf("expected both up and down files for %+v", list[0])
	}
}

func TestCreateMigrationFiles(t *testing.T) {
	dir := t.TempDir()
	ts := time.Date(2026, 11, 21, 14, 30, 5, 0, time.UTC)

	upPath, downPath, err := database.CreateMigrationFiles(dir, "Create Users-Table!", ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPrefix := "20261121143005_create_users_table"
	if filepath.Base(upPath) != expectedPrefix+".up.sql" {
		t.Fatalf("unexpected up filename: %s", filepath.Base(upPath))
	}
	if filepath.Base(downPath) != expectedPrefix+".down.sql" {
		t.Fatalf("unexpected down filename: %s", filepath.Base(downPath))
	}

	upBytes, err := os.ReadFile(upPath)
	if err != nil || !strings.Contains(string(upBytes), expectedPrefix) {
		t.Fatalf("expected up migration template content, err=%v", err)
	}

	loaded, err := database.LoadMigrations(os.DirFS(dir))
	if err != nil {
		t.Fatalf("failed loading created migrations: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Version != expectedPrefix {
		t.Fatalf("unexpected loaded migrations: %+v", loaded)
	}
}
