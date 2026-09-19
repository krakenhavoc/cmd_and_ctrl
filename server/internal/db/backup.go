package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// DefaultBackupInterval is used when CMDCTRL_DB_BACKUP_INTERVAL is
// unset. Hourly is frequent enough that the nightly off-node backup
// (ADR 0051 decision 1, scripts/backup-offsite.sh — not this
// package's job) never picks up a backup more than an hour stale, and
// cheap enough at this project's scale (eight people, one small
// database) that running VACUUM INTO every hour costs nothing anyone
// would notice.
const DefaultBackupInterval = time.Hour

// Backup writes a consistent copy of the live database to
// <dataDir>/db/cmdctrl.backup.sqlite using SQLite's online backup
// primitive, VACUUM INTO — it does not stop or lock out the service
// while it runs. The copy is written to a temp file in the same
// directory and renamed into place, mode 0600, so a concurrent reader
// (the disk-level backup sweep, an operator's sqlite3 session) never
// observes a partially written backup.
func (d *DB) Backup(ctx context.Context) error {
	tmp, err := os.CreateTemp(d.dir, BackupFileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("db: backup temp file: %w", err)
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("db: backup temp file: %w", err)
	}
	// VACUUM INTO refuses to write to a file that already exists, so
	// the temp file is created only to reserve a unique name, then
	// removed before SQLite writes it.
	if err := os.Remove(tmpPath); err != nil {
		return fmt.Errorf("db: backup temp file: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	// VACUUM INTO takes a single connection for its duration; go
	// through Conn rather than sql.DB.ExecContext so a pooled
	// connection isn't recycled mid-statement.
	conn, err := d.DB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("db: backup: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, `VACUUM INTO ?`, tmpPath); err != nil {
		return fmt.Errorf("db: VACUUM INTO: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return fmt.Errorf("db: backup: %w", err)
	}

	finalPath := filepath.Join(d.dir, BackupFileName)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("db: backup: %w", err)
	}
	cleanup = false
	return nil
}

// backupTimeout bounds one Backup call. It is deliberately not the
// loop interval: a short interval (tests, or a small
// CMDCTRL_DB_BACKUP_INTERVAL) would cancel every VACUUM INTO before it
// finished, and the loop would log failures forever. A backup that
// outruns the interval just delays the next tick; the ticker drops
// ticks rather than queueing them.
const backupTimeout = 10 * time.Minute

// RunBackupLoop writes one backup straight away, then calls Backup on
// interval until ctx is canceled. It runs in the caller's goroutine —
// callers start it with `go`. A failed backup is logged and never stops
// the loop or the server: a missed backup is a gap in redundancy, not a
// reason to go down.
//
// The immediate backup is what the off-site copy depends on. Waiting a
// full interval first would leave no consistent copy for that long after
// every restart (and none at all on a fresh host), and the nightly restic
// run takes only the copy, never the live file and its WAL. It also
// captures the schema the boot's migrations just produced.
//
// interval <= 0 disables the loop entirely (it returns immediately,
// writing nothing) rather than defaulting silently, so a
// misconfiguration that meant to disable backups does, and one that
// meant to enable them at the default is spelled out by the caller
// choosing DefaultBackupInterval.
func (d *DB) RunBackupLoop(ctx context.Context, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		return
	}
	d.backupAndLog(ctx, log)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.backupAndLog(ctx, log)
		}
	}
}

// backupAndLog runs one bounded Backup and logs its outcome.
func (d *DB) backupAndLog(ctx context.Context, log *slog.Logger) {
	backupCtx, cancel := context.WithTimeout(ctx, backupTimeout)
	err := d.Backup(backupCtx)
	cancel()
	if err != nil {
		log.Error("database backup failed", "path", d.path, "err", err)
		return
	}
	log.Info("database backup written", "path", filepath.Join(d.dir, BackupFileName))
}
