// Package db is cmd_and_ctrl's persistent store for state that
// outlives a single game: people, their games, their decks and their
// invites (ADR 0051, docs/decisions/0051-user-database.md). Sub-PR 1
// ships only the plumbing this package needs to exist at all — open,
// WAL, migrations, backups — and no business tables. Those land in
// sub-PRs 2 and 3.
//
// The store is a single SQLite file at <dataDir>/db/cmdctrl.sqlite,
// opened by modernc.org/sqlite (pure Go, no cgo, so the binary stays
// static) via database/sql. The game server is its only writer; a
// second process opening this file is exactly the failure ADR 0051
// decision 1 rules out. No ORM, no query builder — hand-written SQL,
// matching the stdlib-first convention elsewhere in this repo.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
)

const (
	// dirName is the subdirectory of the data root the database and
	// its backups live under.
	dirName = "db"
	// fileName is the live database's filename inside dirName.
	fileName = "cmdctrl.sqlite"
	// BackupFileName is the VACUUM INTO backup's filename, written
	// beside the live file so a disk-level backup sweep (HomeLab,
	// nightly, off-node — outside this package) picks up a consistent
	// copy of both from the same directory.
	BackupFileName = "cmdctrl.backup.sqlite"
)

// DB wraps a *sql.DB with the filesystem location the backup sweep
// needs. The embedded *sql.DB means every ordinary database/sql call
// (QueryContext, ExecContext, BeginTx, ...) works directly on a *DB.
type DB struct {
	*sql.DB
	dir  string // <dataDir>/db
	path string // <dataDir>/db/cmdctrl.sqlite
}

// Path returns the live database file's path, for logging.
func (d *DB) Path() string { return d.path }

// Open opens (creating if absent) the SQLite database at
// <dataDir>/db/cmdctrl.sqlite: the directory is created if missing,
// the file is mode 0600 (it will hold invite tokens and, from ADR
// 0051 decision 5, encrypted OAuth refresh tokens), WAL journaling and
// foreign key enforcement are turned on, and a busy timeout absorbs
// the brief contention between the server's own connections. Every
// pending migration is then applied — see migrate().
//
// dataDir must be non-empty; callers that support disabling disk
// persistence entirely (CMDCTRL_DATA_DIR="") decide that before
// calling Open, the same way loadCardAssets and the avatar/bug-report
// stores already do.
//
// A database recording a schema_migrations version newer than this
// binary's embedded migrations returns an error wrapping
// ErrSchemaTooNew — the caller (main) is expected to log it and exit
// rather than run against a schema it does not understand, mirroring
// game.SnapshotSchemaVersion's posture for restore points.
func Open(ctx context.Context, dataDir string) (*DB, error) {
	if dataDir == "" {
		return nil, fmt.Errorf("db: dataDir is empty")
	}
	dir := filepath.Join(dataDir, dirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("db: create %s: %w", dir, err)
	}
	path := filepath.Join(dir, fileName)

	// Touch the file with 0600 before opening it, so a fresh database
	// never has a moment at the default (looser) mode. An existing
	// file is chmod'd too, in case it was created by an older build
	// before this line existed.
	if err := ensureFileMode0600(path); err != nil {
		return nil, fmt.Errorf("db: %s: %w", path, err)
	}

	sqlDB, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db: ping %s: %w", path, err)
	}

	if err := migrate(ctx, sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	// The file mode survives CREATE/write activity, but set it again
	// after migrate() in case SQLite recreated the inode (it does not,
	// normally, but this is cheap insurance against a permissive
	// umask on first create).
	if err := os.Chmod(path, 0o600); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db: chmod %s: %w", path, err)
	}

	return &DB{DB: sqlDB, dir: dir, path: path}, nil
}

// dsn builds the modernc.org/sqlite connection string. Pragmas in the
// query string are applied to every new pooled connection as it opens
// (see applyQueryParams in the driver), which is what makes
// foreign_keys and busy_timeout reliable across the whole pool rather
// than only the first connection.
func dsn(path string) string {
	q := url.Values{}
	// busy_timeout: SQLite's own retry-on-SQLITE_BUSY window. The
	// server is the only writer (ADR 0051 decision 1), so contention
	// is between the server's own goroutines, not a second process —
	// a few seconds is generous rather than load-bearing.
	q.Add("_pragma", "busy_timeout(5000)")
	// WAL: readers don't block the writer and vice versa. This is a
	// database-level setting persisted in the file itself; setting it
	// on every connection is a harmless no-op once it has stuck.
	q.Add("_pragma", "journal_mode(WAL)")
	// Off by default in SQLite for backward compatibility; this
	// project's schema uses foreign keys and expects them enforced.
	q.Add("_pragma", "foreign_keys(1)")
	return "file:" + filepath.ToSlash(path) + "?" + q.Encode()
}

// ensureFileMode0600 creates path if it does not exist (mode 0600) and
// chmods it to 0600 if it does. It never touches the file's contents.
func ensureFileMode0600(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}
