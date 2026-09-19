package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// ErrSchemaTooNew means the database's recorded schema_migrations
// version is higher than any migration this binary embeds. Forward-
// only, no down-migrations: a binary rolled back to an older release
// must refuse a schema it does not understand rather than guess at
// it, the same posture game.SnapshotSchemaVersion takes for restore
// points (server/internal/game/snapshot.go).
var ErrSchemaTooNew = errors.New("db: schema is newer than this binary understands")

// migration is one embedded, numbered SQL file.
type migration struct {
	version int
	name    string
	sql     string
}

// migrate applies every embedded migration newer than the database's
// currently recorded version, each inside its own transaction,
// recording it in schema_migrations as that transaction commits.
// Called once, at boot, before any other package touches the
// database. Idempotent: calling it again with nothing new to apply is
// a single read and no writes.
func migrate(ctx context.Context, sqlDB *sql.DB) error {
	return migrateTo(ctx, sqlDB, 0)
}

// migrateTo is migrate with a ceiling: target > 0 stops after that
// version, so a test can build a database exactly as an older binary
// left it and then run the rest. target <= 0 means "everything".
func migrateTo(ctx context.Context, sqlDB *sql.DB, target int) error {
	migs, err := loadMigrations()
	if err != nil {
		return err
	}
	if len(migs) == 0 {
		return fmt.Errorf("db: no embedded migrations found")
	}
	maxKnown := migs[len(migs)-1].version

	current, err := currentSchemaVersion(ctx, sqlDB)
	if err != nil {
		return err
	}
	if current > maxKnown {
		return fmt.Errorf("%w: database is at v%d, this binary knows up to v%d",
			ErrSchemaTooNew, current, maxKnown)
	}

	for _, m := range migs {
		if m.version <= current {
			continue
		}
		if target > 0 && m.version > target {
			break
		}
		if err := applyMigration(ctx, sqlDB, m); err != nil {
			return fmt.Errorf("db: migration %04d_%s: %w", m.version, m.name, err)
		}
	}
	return nil
}

// currentSchemaVersion reads the highest applied migration version.
// A database with no schema_migrations table yet — the very first
// boot, before migration 0001 has run — reads as version 0, not an
// error.
func currentSchemaVersion(ctx context.Context, sqlDB *sql.DB) (int, error) {
	var version sql.NullInt64
	err := sqlDB.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&version)
	if err != nil {
		if isNoSuchTable(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("db: read schema_migrations: %w", err)
	}
	return int(version.Int64), nil
}

// isNoSuchTable reports whether err is SQLite's "no such table" error.
// modernc.org/sqlite surfaces this as a plain error string rather
// than a typed sentinel, so a substring check is the only option —
// the same thing every other SQLite-backed Go project does for this
// case.
func isNoSuchTable(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such table")
}

// applyMigration runs one migration's SQL and records it, atomically:
// either both happen or neither does, so a crash mid-migration never
// leaves schema_migrations disagreeing with the schema it describes.
//
// # Foreign keys are off while a migration runs
//
// Every pooled connection has foreign_keys on (see dsn). A migration
// runs on one pinned connection with it switched OFF for the duration,
// and checks the result with PRAGMA foreign_key_check before it
// commits. This is SQLite's documented procedure for changing a table
// definition (https://sqlite.org/lang_altertable.html#otheralter), and
// it is not optional once a migration rebuilds a parent table:
//
//   - PRAGMA foreign_keys cannot be changed inside a transaction — it
//     is a silent no-op there — so it has to be set on the connection
//     BEFORE BEGIN, which is why the connection is pinned.
//   - With it on, DROP TABLE on a parent does an implicit DELETE FROM
//     that fires ON DELETE CASCADE on its children: rebuilding `games`
//     would delete every seat and invite. PRAGMA defer_foreign_keys
//     does not help; it defers the constraint check, not the cascade.
//   - With it on, ALTER TABLE ... RENAME rewrites the REFERENCES
//     clauses of other tables to follow the renamed table.
//
// The check before COMMIT is what keeps "off" honest: a migration that
// leaves a dangling reference fails and rolls back, exactly as if the
// constraint had been enforced row by row.
//
// The connection goes back to the pool with foreign_keys on again. If
// switching it back fails, the connection is discarded rather than
// returned, so no later query can run on it with enforcement off.
func applyMigration(ctx context.Context, sqlDB *sql.DB, m migration) (err error) {
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}
	defer func() {
		// A fresh context: ctx being cancelled may be why we are
		// unwinding, and this must still run.
		if _, rerr := conn.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`); rerr != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
			if err == nil {
				err = fmt.Errorf("re-enable foreign keys: %w", rerr)
			}
		}
	}()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful Commit

	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return err
	}
	if err := foreignKeyCheck(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
		m.version, m.name, time.Now().Unix()); err != nil {
		return err
	}
	return tx.Commit()
}

// foreignKeyCheck runs PRAGMA foreign_key_check over the whole
// database and turns any violation into an error naming the first few.
func foreignKeyCheck(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("foreign_key_check: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var (
		violations []string
		total      int
	)
	for rows.Next() {
		var (
			table, parent string
			rowid         sql.NullInt64
			fkid          int
		)
		if err := rows.Scan(&table, &rowid, &parent, &fkid); err != nil {
			return fmt.Errorf("foreign_key_check: %w", err)
		}
		total++
		if len(violations) < 5 {
			violations = append(violations, fmt.Sprintf("%s rowid %d -> %s", table, rowid.Int64, parent))
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("foreign_key_check: %w", err)
	}
	if total > 0 {
		return fmt.Errorf("migration leaves %d dangling foreign key(s): %s", total, strings.Join(violations, "; "))
	}
	return nil
}

// loadMigrations reads every embedded *.sql file, parses its version
// from the numeric filename prefix, and returns them sorted ascending.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("db: read embedded migrations: %w", err)
	}

	migs := make([]migration, 0, len(entries))
	seen := make(map[int]string, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		version, name, err := parseMigrationFilename(e.Name())
		if err != nil {
			return nil, fmt.Errorf("db: %s: %w", e.Name(), err)
		}
		if prev, ok := seen[version]; ok {
			return nil, fmt.Errorf("db: migration version %d declared twice (%s and %s)", version, prev, e.Name())
		}
		seen[version] = e.Name()

		raw, err := migrationFiles.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("db: read %s: %w", e.Name(), err)
		}
		migs = append(migs, migration{version: version, name: name, sql: string(raw)})
	}
	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })
	return migs, nil
}

// parseMigrationFilename splits "0001_schema_migrations.sql" into its
// version (1) and name ("schema_migrations").
func parseMigrationFilename(name string) (int, string, error) {
	base := strings.TrimSuffix(name, ".sql")
	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("expected NNNN_name.sql")
	}
	version, err := strconv.Atoi(parts[0])
	if err != nil || version <= 0 {
		return 0, "", fmt.Errorf("expected a positive numeric prefix: %w", err)
	}
	return version, parts[1], nil
}
