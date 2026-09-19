package lobby

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// persist_test.go covers the lobby half of a restart: the game came
// back, but can anyone get INTO it?
//
// A player whose session did not survive (no CMDCTRL_SESSION_KEY, or
// a lost or expired token) re-authenticates through the invite link.
// If the invite were not persisted, a restored game would be a table
// nobody can open — which is indistinguishable, from the player's
// seat, from having lost the game. Since ADR 0051 decision 4 the
// lobby's half lives in the database as games / seats / invites rows.

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// openTestDB opens (and migrates) the database under dir, closed at
// the end of the test. Calling it twice on one dir is the deploy: a
// second process opening the same file.
func openTestDB(t *testing.T, dir string) *db.DB {
	t.Helper()
	d, err := db.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

// newDurableLobby is one server process with a data dir and a
// database: what production runs. Tests that restart build a second
// one over the same dir.
func newDurableLobby(t *testing.T, dir string) (*Lobby, *ws.RoomManager) {
	t.Helper()
	mgr := ws.NewRoomManager(quietLogger(), dir)
	return NewLobbyWithStore(mgr, NewSQLStore(openTestDB(t, dir))), mgr
}

// TestInviteSurvivesARestart is the test that makes the feature
// usable rather than merely correct.
func TestInviteSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Friday Night Commander")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join alice: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Bob"); err != nil {
		t.Fatalf("Join bob: %v", err)
	}

	// --- the deploy -----------------------------------------------
	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(log); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}

	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatalf("restored lobby does not know the game: %v", err)
	}
	if back.Name != "Friday Night Commander" {
		t.Errorf("name = %q, want the original", back.Name)
	}
	if !back.CreatedAt.Equal(meta.CreatedAt.Truncate(1e6)) {
		t.Errorf("created_at = %v, want %v to the millisecond", back.CreatedAt, meta.CreatedAt)
	}
	if len(back.Players) != 2 {
		t.Fatalf("seat list has %d entries, want 2", len(back.Players))
	}
	if back.Players[0].Name != "Alice" || back.Players[1].Name != "Bob" {
		t.Errorf("seat list = %+v, want Alice then Bob", back.Players)
	}
	// Only the hashes were stored, so the new process cannot show the
	// tokens — but the links must still work.
	if back.InviteToken != "" || back.SpectatorInvite != "" {
		t.Errorf("restored meta shows invite tokens the database never held: %q / %q",
			back.InviteToken, back.SpectatorInvite)
	}
	if _, _, err := l2.Join(back.ID, meta.InviteToken, "Carol"); err != nil {
		t.Errorf("restored player invite was rejected: %v", err)
	}
	if _, err := l2.Spectate(back.ID, meta.SpectatorInvite); err != nil {
		t.Errorf("restored spectator invite was rejected: %v", err)
	}
	if id, err := l2.FindByInvite(meta.InviteToken); err != nil || id != meta.ID {
		t.Errorf("FindByInvite after restart = %v, %v; want %v", id, err, meta.ID)
	}
}

// TestInviteTokensAreStoredOnlyAsHashes — the database is backed up
// and copied around; a plaintext token in it would be a live
// credential in every copy.
func TestInviteTokensAreStoredOnlyAsHashes(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Secrets")
	if err != nil {
		t.Fatal(err)
	}
	d := openTestDB(t, dir)
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM invites WHERE game_id = ?`, meta.ID.String()).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("invites rows = %d, want 2", n)
	}
	for _, tok := range []string{meta.InviteToken, meta.SpectatorInvite} {
		h, ok := hashInvite(tok)
		if !ok {
			t.Fatalf("minted token %q does not hash", tok)
		}
		var kind string
		if err := d.QueryRow(`SELECT kind FROM invites WHERE token_hash = ?`, h[:]).Scan(&kind); err != nil {
			t.Errorf("no row for the hash of %q: %v", tok, err)
		}
	}
	// And no copy of the plaintext anywhere in the file set.
	for _, name := range []string{"cmdctrl.sqlite", "cmdctrl.sqlite-wal"} {
		raw, err := os.ReadFile(filepath.Join(dir, "db", name))
		if err != nil {
			continue
		}
		for _, tok := range []string{meta.InviteToken, meta.SpectatorInvite} {
			if bytes.Contains(raw, []byte(tok)) {
				t.Errorf("%s contains a plaintext invite token", name)
			}
		}
	}
}

// TestGameWithoutMetadataIsDropped: a restored engine snapshot whose
// games row is missing is not a usable game — nobody can be invited to
// it — so it must not be left registered and half-visible.
func TestGameWithoutMetadataIsDropped(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Doomed")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join: %v", err)
	}

	// Simulate the metadata being lost while the engine snapshot
	// survived.
	if err := NewSQLStore(openTestDB(t, dir)).DeleteGame(context.Background(), meta.ID); err != nil {
		t.Fatalf("delete rows: %v", err)
	}

	l2, mgr2 := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(log); n != 0 {
		t.Errorf("restored %d games, want 0", n)
	}
	if _, err := l2.Get(meta.ID); err == nil {
		t.Error("a game nobody can be invited to was left in the lobby")
	}
	if mgr2.Get(meta.ID) != nil {
		t.Error("the orphaned room was left registered with the manager")
	}
}

// TestDeletedGameStaysDeleted — the rows must go with the game, or the
// next boot resurrects something the operator removed, and its invite
// links keep resolving.
func TestDeletedGameStaysDeleted(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Transient")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join: %v", err)
	}
	// An importer leftover for the same game is reaped too.
	legacy := legacyMetaPath(dir, meta.ID) + importedSuffix
	if err := os.MkdirAll(filepath.Dir(legacy), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := l.Delete(meta.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	d := openTestDB(t, dir)
	for _, q := range []string{
		`SELECT COUNT(*) FROM games WHERE id = ?`,
		`SELECT COUNT(*) FROM seats WHERE game_id = ?`,
		`SELECT COUNT(*) FROM invites WHERE game_id = ?`,
	} {
		var n int
		if err := d.QueryRow(q, meta.ID.String()).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("%s = %d after Delete, want 0", q, n)
		}
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("imported legacy file survived Delete (err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "restore", meta.ID.String()+".json")); !os.IsNotExist(err) {
		t.Errorf("restore point survived Delete (err=%v)", err)
	}
	if _, err := l.FindByInvite(meta.InviteToken); err != ErrInvalidInvite {
		t.Errorf("a deleted game's invite still resolves: %v", err)
	}

	l2, _ := newDurableLobby(t, dir)
	l2.RestoreFromDisk(log)
	if _, err := l2.Get(meta.ID); err == nil {
		t.Error("a deleted game came back after a restart")
	}
}

// TestArchiveIsAColumn — archiving writes archived_at and leaves every
// row and file where it was; unarchiving clears it.
func TestArchiveIsAColumn(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _, _ := startTwoSeatGame(t, l, "Retire me")
	d := openTestDB(t, dir)

	archivedAt := func() (valid bool) {
		t.Helper()
		var at *int64
		if err := d.QueryRow(`SELECT archived_at FROM games WHERE id = ?`, meta.ID.String()).Scan(&at); err != nil {
			t.Fatal(err)
		}
		return at != nil
	}

	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if !archivedAt() {
		t.Error("archived_at not set by SetArchived(true)")
	}
	var seats, invites int
	_ = d.QueryRow(`SELECT COUNT(*) FROM seats WHERE game_id = ?`, meta.ID.String()).Scan(&seats)
	_ = d.QueryRow(`SELECT COUNT(*) FROM invites WHERE game_id = ?`, meta.ID.String()).Scan(&invites)
	if seats != 2 || invites != 2 {
		t.Errorf("archive touched rows: %d seats, %d invites; want 2 and 2", seats, invites)
	}
	if _, err := os.Stat(filepath.Join(dir, "restore", meta.ID.String()+".json")); err != nil {
		t.Errorf("archive removed the restore point: %v", err)
	}

	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	if archivedAt() {
		t.Error("archived_at still set after SetArchived(false)")
	}
}

// TestLifecycleColumnsFollowTheGame — state, started_at, ended_at and
// winner_seat are kept current from transitions the lobby sees,
// including a game that ends over the WebSocket with nobody looking at
// the lobby.
func TestLifecycleColumnsFollowTheGame(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, alice, _ := startTwoSeatGame(t, l, "Lifecycle")
	d := openTestDB(t, dir)

	read := func() (state string, started, ended, winner *int64) {
		t.Helper()
		if err := d.QueryRow(`SELECT state, started_at, ended_at, winner_seat FROM games WHERE id = ?`,
			meta.ID.String()).Scan(&state, &started, &ended, &winner); err != nil {
			t.Fatal(err)
		}
		return
	}
	state, started, ended, _ := read()
	if state != "active" || started == nil || ended != nil {
		t.Fatalf("after Start: state=%q started=%v ended=%v", state, started, ended)
	}

	// Alice concedes through the room, the way a WS action does. The
	// lobby is not called.
	room := l.RoomOf(meta.ID)
	if _, _, err := room.Apply(alice, func() error { return room.Game.Concede(alice) }); err != nil {
		t.Fatalf("concede: %v", err)
	}
	waitFor(t, func() bool {
		s, _, e, w := read()
		return s == "ended" && e != nil && w != nil
	})
	_, _, _, winner := read()
	if *winner != 1 {
		t.Errorf("winner_seat = %d, want 1 (Bob; Alice conceded)", *winner)
	}
}

// TestNoDatabaseStillPlays — CMDCTRL_DATA_DIR="" means no database;
// the lobby runs on the memory store and every invite path works in
// process. Nothing is restored, and nothing on disk is deleted.
func TestNoDatabaseStillPlays(t *testing.T) {
	mgr := ws.NewRoomManager(quietLogger(), "")
	l := NewLobby(mgr)
	meta, err := l.Create("No disk")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if _, err := l.Spectate(meta.ID, meta.SpectatorInvite); err != nil {
		t.Fatalf("Spectate: %v", err)
	}
	if id, err := l.FindByInvite(meta.InviteToken); err != nil || id != meta.ID {
		t.Fatalf("FindByInvite = %v, %v", id, err)
	}
	if _, kind, err := l.Preview(meta.ID, meta.SpectatorInvite); err != nil || kind != PreviewSpectator {
		t.Fatalf("Preview = %v, %v", kind, err)
	}
	if n := l.RestoreFromDisk(quietLogger()); n != 0 {
		t.Errorf("restored %d games with no data dir", n)
	}
	if err := l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := l.FindByInvite(meta.InviteToken); err != ErrInvalidInvite {
		t.Errorf("deleted game's invite still resolves in memory: %v", err)
	}
}

// A data dir with no database (only reachable in tests; main opens
// one whenever the dir is set) must not treat every restore point as
// orphaned and delete it.
func TestRestoreWithoutADurableStoreLeavesRestorePointsAlone(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _, _ := startTwoSeatGame(t, l, "Keep me")

	mgr2 := ws.NewRoomManager(quietLogger(), dir)
	l2 := NewLobby(mgr2)
	if n := l2.RestoreFromDisk(quietLogger()); n != 0 {
		t.Errorf("restored %d games from a memory store", n)
	}
	if _, err := os.Stat(filepath.Join(dir, "restore", meta.ID.String()+".json")); err != nil {
		t.Errorf("restore point was removed: %v", err)
	}
}

// waitFor polls cond until it holds or five seconds pass.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met within 5s")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
