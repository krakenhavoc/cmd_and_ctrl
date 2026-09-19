package db

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestBackupProducesAReadableDatabase(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	if _, err := d.ExecContext(ctx, `CREATE TABLE probe (id INTEGER PRIMARY KEY, note TEXT NOT NULL)`); err != nil {
		t.Fatalf("create probe table: %v", err)
	}
	if _, err := d.ExecContext(ctx, `INSERT INTO probe (id, note) VALUES (1, 'hello')`); err != nil {
		t.Fatalf("insert probe row: %v", err)
	}

	if err := d.Backup(ctx); err != nil {
		t.Fatalf("Backup: %v", err)
	}

	backupPath := filepath.Join(dir, dirName, BackupFileName)
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("stat backup file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("backup file mode = %04o, want 0600", perm)
	}

	// Open the backup as an independent database, through the driver
	// directly rather than db.Open, so this test proves the FILE is a
	// valid, complete SQLite database — not merely that our own Open
	// can read it.
	backupDB, err := sql.Open("sqlite", "file:"+filepath.ToSlash(backupPath)+"?mode=ro")
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	defer backupDB.Close()

	var note string
	if err := backupDB.QueryRowContext(ctx, `SELECT note FROM probe WHERE id = 1`).Scan(&note); err != nil {
		t.Fatalf("query backup: %v", err)
	}
	if note != "hello" {
		t.Errorf("backup probe.note = %q, want %q", note, "hello")
	}

	var migCount int
	if err := backupDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&migCount); err != nil {
		t.Fatalf("query backup schema_migrations: %v", err)
	}
	if migCount == 0 {
		t.Error("backup schema_migrations is empty; VACUUM INTO should have copied it")
	}
}

func TestBackupOverwritesPreviousBackup(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	if err := d.Backup(ctx); err != nil {
		t.Fatalf("first Backup: %v", err)
	}
	if _, err := d.ExecContext(ctx, `CREATE TABLE probe (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create probe table: %v", err)
	}
	if err := d.Backup(ctx); err != nil {
		t.Fatalf("second Backup: %v", err)
	}

	backupPath := filepath.Join(dir, dirName, BackupFileName)
	backupDB, err := sql.Open("sqlite", "file:"+filepath.ToSlash(backupPath)+"?mode=ro")
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	defer backupDB.Close()

	var name string
	err = backupDB.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'probe'`).Scan(&name)
	if err != nil {
		t.Fatalf("second backup does not contain the table created after the first backup: %v", err)
	}
}

func TestRunBackupLoopRunsOnInterval(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1})) // discard

	done := make(chan struct{})
	go func() {
		d.RunBackupLoop(ctx, log, 20*time.Millisecond)
		close(done)
	}()

	backupPath := filepath.Join(dir, dirName, BackupFileName)
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(backupPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("RunBackupLoop did not produce a backup within the deadline")
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunBackupLoop did not return after context cancellation")
	}
}

func TestRunBackupLoopBacksUpImmediately(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1})) // discard

	done := make(chan struct{})
	go func() {
		// An hour: the first tick cannot fire during this test, so a
		// backup that appears came from the immediate one at start.
		d.RunBackupLoop(ctx, log, time.Hour)
		close(done)
	}()

	backupPath := filepath.Join(dir, dirName, BackupFileName)
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(backupPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("RunBackupLoop did not write a backup at start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("RunBackupLoop did not return after context cancellation")
	}
}

func TestRunBackupLoopDisabledWithNonPositiveInterval(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	d, err := Open(ctx, dir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError + 1}))

	done := make(chan struct{})
	go func() {
		d.RunBackupLoop(ctx, log, 0)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("RunBackupLoop with interval <= 0 did not return immediately")
	}
	if _, err := os.Stat(filepath.Join(dir, dirName, BackupFileName)); !os.IsNotExist(err) {
		t.Fatalf("RunBackupLoop with interval <= 0 wrote a backup (stat err: %v)", err)
	}
}
