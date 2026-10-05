package db

import (
	"context"
	"testing"
	"time"
)

// LastBackup and SizeBytes back cmdctrl_db_backup_last_success_timestamp_seconds
// and cmdctrl_db_size_bytes (ADR 0123 §3).
func TestLastBackupAndSize(t *testing.T) {
	ctx := context.Background()
	d, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	if !d.LastBackup().IsZero() {
		t.Errorf("LastBackup before any backup = %v, want zero", d.LastBackup())
	}
	n, err := d.SizeBytes()
	if err != nil || n <= 0 {
		t.Errorf("SizeBytes = %d, %v; want a migrated file's size", n, err)
	}

	before := time.Now().Add(-time.Millisecond)
	if err := d.Backup(ctx); err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if at := d.LastBackup(); at.Before(before) || at.After(time.Now()) {
		t.Errorf("LastBackup after a backup = %v, want about now", at)
	}

	// A failed backup leaves the last success alone.
	last := d.LastBackup()
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := d.Backup(canceled); err == nil {
		t.Fatal("Backup on a canceled context succeeded")
	}
	if !d.LastBackup().Equal(last) {
		t.Errorf("a failed backup moved LastBackup from %v to %v", last, d.LastBackup())
	}
}
