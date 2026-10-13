package lobby

// unresident_test.go covers #2987: a games row the lobby does not hold
// (the restore pass never brought its room back) must stop reading as
// active, and an admin must be able to delete or archive it.

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func restorePointFor(dir string, id uuid.UUID) string {
	return filepath.Join(dir, "restore", id.String()+".json")
}

func storedState(t *testing.T, dir string, id uuid.UUID) GameRecord {
	t.Helper()
	rec, _, err := NewSQLStore(openTestDB(t, dir)).LoadGame(context.Background(), id)
	if err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	return rec
}

// seedUnresident writes a games row straight into l's store, with no
// room behind it: the shape the prod load test left behind.
func seedUnresident(t *testing.T, l *Lobby, name, state string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	rec := GameRecord{ID: id, Name: name, State: state, CreatedAt: time.Now().UTC()}
	if err := l.store.CreateGame(context.Background(), rec, nil); err != nil {
		t.Fatalf("seed row: %v", err)
	}
	if err := l.store.ReplaceSeats(context.Background(), id, []SeatRecord{
		{Seat: 0, PlayerID: uuid.New(), GuestName: "LT 1", BotTier: "heuristic"},
	}); err != nil {
		t.Fatalf("seed seats: %v", err)
	}
	return id
}

// A restore point that cannot be decoded leaves the table out of the
// lobby. Its row must not keep saying active.
func TestFailedRestoreLeavesTheRowEnded(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _, _ := startTwoSeatGame(t, l, "Doomed")
	if err := os.WriteFile(restorePointFor(dir, meta.ID), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 0 {
		t.Fatalf("restored %d games, want 0", n)
	}
	rec := storedState(t, dir, meta.ID)
	if rec.State != string(game.StateEnded) {
		t.Errorf("row state = %q after a failed restore, want ended", rec.State)
	}
	if rec.EndedAt == nil {
		t.Error("closed row has no ended_at")
	}
	if rec.Outcome != "" {
		t.Errorf("closed row outcome = %q, want none", rec.Outcome)
	}
}

// The prod case: the process died before the table's first restore
// point reached disk, so RestoreRooms never saw it at all.
func TestRowWithNoRestorePointIsEnded(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _, _ := startTwoSeatGame(t, l, "Never snapshotted")
	if err := os.Remove(restorePointFor(dir, meta.ID)); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	l2.RestoreFromDisk(quietLogger())
	if got := storedState(t, dir, meta.ID).State; got != string(game.StateEnded) {
		t.Errorf("row state = %q with no restore point, want ended", got)
	}
}

// A table that did come back is untouched, and so is one whose
// restore point was left for a roll-forward.
func TestRestoredAndRollForwardRowsAreLeftAlone(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	live, _, _ := startTwoSeatGame(t, l, "Comes back")
	newer, _, _ := startTwoSeatGame(t, l, "Written by a newer build")

	path := restorePointFor(dir, newer.ID)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file map[string]any
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	file["snapshot"].(map[string]any)["schema"] = game.SnapshotSchemaVersion + 1
	bumped, _ := json.Marshal(file)
	if err := os.WriteFile(path, bumped, 0o644); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	if got := storedState(t, dir, live.ID).State; got != string(game.StateActive) {
		t.Errorf("restored row state = %q, want active", got)
	}
	if got := storedState(t, dir, newer.ID).State; got != string(game.StateActive) {
		t.Errorf("roll-forward row state = %q, want active (file is kept for the newer build)", got)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("roll-forward restore point was removed: %v", err)
	}
}

// A finished game's row is history, not an orphan.
func TestEndedRowStaysAsItWas(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	id := seedUnresident(t, l, "History", string(game.StateEnded))
	ended := time.Now().Add(-48 * time.Hour).UTC().Truncate(time.Millisecond)
	rec := storedState(t, dir, id)
	rec.EndedAt = &ended
	rec.Outcome = "win"
	if err := l.store.UpdateGame(context.Background(), rec); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	l2.RestoreFromDisk(quietLogger())
	got := storedState(t, dir, id)
	if got.Outcome != "win" || got.EndedAt == nil || !got.EndedAt.Equal(ended) {
		t.Errorf("ended row was rewritten: %+v", got)
	}
}

func TestDeleteReachesARowTheLobbyDoesNotHold(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	id := seedUnresident(t, l, "loadtest-orphan", string(game.StateActive))
	// A restore point left on disk for it must go too.
	if err := os.MkdirAll(filepath.Join(dir, "restore"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(restorePointFor(dir, id), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := l.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, _, err := l.store.LoadGame(context.Background(), id); err == nil {
		t.Error("row survived Delete")
	}
	if _, err := os.Stat(restorePointFor(dir, id)); !os.IsNotExist(err) {
		t.Errorf("restore point survived Delete (err=%v)", err)
	}
	if err := l.Delete(id); err != ErrGameNotFound {
		t.Errorf("second Delete = %v, want ErrGameNotFound", err)
	}
	if err := l.Delete(uuid.New()); err != ErrGameNotFound {
		t.Errorf("Delete of an unknown id = %v, want ErrGameNotFound", err)
	}
}

func TestArchiveReachesARowTheLobbyDoesNotHold(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	id := seedUnresident(t, l, "loadtest-orphan", string(game.StateActive))

	meta, err := l.SetArchived(id, true)
	if err != nil {
		t.Fatalf("SetArchived: %v", err)
	}
	if !meta.Archived() || meta.ID != id || len(meta.Players) != 1 {
		t.Errorf("meta = %+v", meta)
	}
	if storedState(t, dir, id).ArchivedAt == nil {
		t.Error("archived_at not written")
	}
	if _, err := l.SetArchived(id, false); err != nil {
		t.Fatalf("unarchive: %v", err)
	}
	if storedState(t, dir, id).ArchivedAt != nil {
		t.Error("archived_at not cleared")
	}
	if _, err := l.SetArchived(uuid.New(), true); err != ErrGameNotFound {
		t.Errorf("SetArchived of an unknown id = %v, want ErrGameNotFound", err)
	}
}

// The routes stay admin-only for a DB-only row, and answer for it.
func TestDeleteAndArchiveOfUnresidentRowOverHTTP(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	adminTok := adminSession(t, a)
	id := seedUnresident(t, l, "loadtest-orphan", string(game.StateActive))
	playerTok := playerSession(t, a, id)
	path := "/games/" + id.String()

	call := func(method, p, tok string) int {
		t.Helper()
		req, _ := http.NewRequest(method, srv.URL+p, nil)
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}

	for _, c := range []struct{ method, path string }{
		{http.MethodDelete, path},
		{http.MethodPost, path + "/archive"},
		{http.MethodDelete, path + "/archive"},
	} {
		if got := call(c.method, c.path, ""); got != http.StatusUnauthorized {
			t.Errorf("anonymous %s %s = %d, want 401", c.method, c.path, got)
		}
		if got := call(c.method, c.path, playerTok); got != http.StatusForbidden {
			t.Errorf("player %s %s = %d, want 403", c.method, c.path, got)
		}
	}
	if _, _, err := l.store.LoadGame(context.Background(), id); err != nil {
		t.Fatalf("a refused call touched the row: %v", err)
	}

	if got := call(http.MethodPost, path+"/archive", adminTok); got != http.StatusOK {
		t.Errorf("admin archive = %d, want 200", got)
	}
	if got := call(http.MethodDelete, path+"/archive", adminTok); got != http.StatusOK {
		t.Errorf("admin unarchive = %d, want 200", got)
	}
	if got := call(http.MethodDelete, path, adminTok); got != http.StatusNoContent {
		t.Errorf("admin delete = %d, want 204", got)
	}
	if got := call(http.MethodDelete, path, adminTok); got != http.StatusNotFound {
		t.Errorf("admin delete of a gone row = %d, want 404", got)
	}
	if got := call(http.MethodPost, "/games/"+uuid.NewString()+"/archive", adminTok); got != http.StatusNotFound {
		t.Errorf("admin archive of an unknown id = %d, want 404", got)
	}
}
