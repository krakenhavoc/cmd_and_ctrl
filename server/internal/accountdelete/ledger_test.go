package accountdelete

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func TestLedgerAppendReadDedupes(t *testing.T) {
	dir := t.TempDir()
	l := NewLedger(dir)
	a, b := uuid.New(), uuid.New()
	at := time.UnixMilli(1700000000000)
	for _, id := range []uuid.UUID{a, b, a} {
		if err := l.Append(id, at); err != nil {
			t.Fatal(err)
		}
	}
	got, err := ReadLedger(LedgerPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].UserID != a || got[1].UserID != b || got[0].DeletedAt != 1700000000000 {
		t.Errorf("ReadLedger = %+v", got)
	}
	info, err := os.Stat(LedgerPath(dir))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("ledger mode = %v, %v; want 0600", info.Mode().Perm(), err)
	}
}

func TestReadLedgerMissingAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	if got, err := ReadLedger(LedgerPath(dir)); err != nil || got != nil {
		t.Errorf("missing ledger = %v, %v; want empty", got, err)
	}
	p := filepath.Join(dir, "bad.jsonl")
	if err := os.WriteFile(p, []byte("{\"user_id\":\"not a uuid\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadLedger(p); err == nil || !strings.Contains(err.Error(), "line 1") {
		t.Errorf("corrupt ledger error = %v, want one naming line 1", err)
	}
}

func TestDeleteRecordsInLedgerOnlyWhenCommitted(t *testing.T) {
	f := seed(t)
	dir := t.TempDir()
	s := NewSQLStore(f.d).WithLedger(NewLedger(dir))
	ctx := context.Background()

	if _, err := s.Delete(ctx, uuid.New()); err == nil {
		t.Fatal("deleting an unknown user succeeded")
	}
	if got, _ := ReadLedger(LedgerPath(dir)); len(got) != 0 {
		t.Fatalf("a failed deletion reached the ledger: %+v", got)
	}
	if _, err := s.Delete(ctx, f.user); err != nil {
		t.Fatal(err)
	}
	got, err := ReadLedger(LedgerPath(dir))
	if err != nil || len(got) != 1 || got[0].UserID != f.user {
		t.Fatalf("ledger = %+v, %v; want the deleted user", got, err)
	}
}

// A restore point taken before the deletion brings the account back;
// re-applying the ledger deletes it again, even where the restored copy
// shows the person at a table that has not ended.
func TestReapplyLedgerAfterRestoringAnOlderCopy(t *testing.T) {
	f := seed(t)
	ctx := context.Background()
	if err := f.d.Backup(ctx); err != nil {
		t.Fatal(err)
	}
	ledgerDir := t.TempDir()
	if _, err := NewSQLStore(f.d).WithLedger(NewLedger(ledgerDir)).Delete(ctx, f.user); err != nil {
		t.Fatal(err)
	}

	// "Restore": the older copy becomes the live database of a fresh data dir.
	restored := t.TempDir()
	if err := os.MkdirAll(filepath.Join(restored, "db"), 0o700); err != nil {
		t.Fatal(err)
	}
	copyTo(t, filepath.Join(filepath.Dir(f.d.Path()), db.BackupFileName), filepath.Join(restored, "db", "cmdctrl.sqlite"))
	rd, err := db.Open(ctx, restored)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rd.Close() }()
	if n := count(t, rd, `SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, f.user.String()); n != 1 {
		t.Fatalf("the restored copy should hold the account live, got %d", n)
	}
	// Make the person look seated at a live table, as an old copy may.
	exec(t, rd, `UPDATE games SET state = 'active', ended_at = NULL WHERE id = ?`, f.endedGame.String())

	entries, err := ReadLedger(LedgerPath(ledgerDir))
	if err != nil {
		t.Fatal(err)
	}
	s := NewSQLStore(rd)
	dry, err := ReapplyLedger(ctx, s, entries, true)
	if err != nil || dry.Applied != 1 {
		t.Fatalf("dry run = %+v, %v", dry, err)
	}
	if n := count(t, rd, `SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NULL`, f.user.String()); n != 1 {
		t.Fatal("a dry run changed the database")
	}
	rep, err := ReapplyLedger(ctx, s, entries, false)
	if err != nil || rep.Applied != 1 || len(rep.PlaymatIDs) != 1 {
		t.Fatalf("reapply = %+v, %v", rep, err)
	}
	if n := count(t, rd, `SELECT COUNT(*) FROM users WHERE id = ? AND deleted_at IS NOT NULL AND display_name = ?`,
		f.user.String(), DeletedPlayerName); n != 1 {
		t.Error("the account is not a tombstone after re-applying")
	}
	if n := count(t, rd, `SELECT COUNT(*) FROM decks WHERE owner_id = ?`, f.user.String()); n != 0 {
		t.Errorf("decks left: %d", n)
	}
	// Idempotent.
	again, err := ReapplyLedger(ctx, s, entries, false)
	if err != nil || again.Applied != 0 || again.AlreadyDeleted != 1 {
		t.Errorf("second reapply = %+v, %v", again, err)
	}
	// An account the older copy never had is not an error.
	unk, err := ReapplyLedger(ctx, s, []LedgerEntry{{UserID: uuid.New()}}, false)
	if err != nil || unk.Unknown != 1 {
		t.Errorf("unknown = %+v, %v", unk, err)
	}
}

func copyTo(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		t.Fatal(err)
	}
}
