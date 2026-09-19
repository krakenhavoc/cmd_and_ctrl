package db

import (
	"context"
	"testing"
)

func TestMigrationsAppliedOnce(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}

	var count int
	if err := d.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != len(migs) {
		t.Errorf("schema_migrations has %d rows, want %d (one per embedded migration)", count, len(migs))
	}

	current, err := currentSchemaVersion(ctx, d.DB)
	if err != nil {
		t.Fatalf("currentSchemaVersion: %v", err)
	}
	if current != migs[len(migs)-1].version {
		t.Errorf("currentSchemaVersion = %d, want %d", current, migs[len(migs)-1].version)
	}
}

// TestMigrationsAreIdempotent reopens the same database and confirms
// migrate() runs cleanly a second time (via Open) without trying to
// re-apply anything or erroring on the already-created table.
func TestMigrationsAreIdempotent(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d1, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if err := d1.Close(); err != nil {
		t.Fatalf("close first handle: %v", err)
	}

	d2, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("second Open (idempotent migrate): %v", err)
	}
	defer d2.Close()

	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	var count int
	if err := d2.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != len(migs) {
		t.Errorf("schema_migrations has %d rows after a second Open, want %d (no re-apply, no duplicates)", count, len(migs))
	}
}

func TestParseMigrationFilename(t *testing.T) {
	cases := []struct {
		name        string
		wantVersion int
		wantName    string
		wantErr     bool
	}{
		{"0001_schema_migrations.sql", 1, "schema_migrations", false},
		{"0042_users_and_identities.sql", 42, "users_and_identities", false},
		{"no_version.sql", 0, "", true},
		{"0001.sql", 0, "", true},
		{"abc_name.sql", 0, "", true},
		{"0000_zero_version.sql", 0, "", true},
	}
	for _, tc := range cases {
		version, name, err := parseMigrationFilename(tc.name)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseMigrationFilename(%q): want error, got version=%d name=%q", tc.name, version, name)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseMigrationFilename(%q): unexpected error: %v", tc.name, err)
			continue
		}
		if version != tc.wantVersion || name != tc.wantName {
			t.Errorf("parseMigrationFilename(%q) = (%d, %q), want (%d, %q)", tc.name, version, name, tc.wantVersion, tc.wantName)
		}
	}
}

func TestLoadMigrationsSortedAscending(t *testing.T) {
	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	if len(migs) == 0 {
		t.Fatal("no embedded migrations")
	}
	for i := 1; i < len(migs); i++ {
		if migs[i-1].version >= migs[i].version {
			t.Errorf("migrations not strictly ascending at index %d: %d >= %d", i, migs[i-1].version, migs[i].version)
		}
	}
}
