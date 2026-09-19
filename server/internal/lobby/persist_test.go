package lobby

import (
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// persist_test.go covers the lobby half of a restart: the game came
// back, but can anyone get INTO it?
//
// A player whose session did not survive (no CMDCTRL_SESSION_KEY, or
// a lost or expired token) re-authenticates through the invite link. If the invite token were not persisted, a restored
// game would be a table nobody can open — which is indistinguishable,
// from the player's seat, from having lost the game.

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// TestInviteSurvivesARestart is the test that makes the feature
// usable rather than merely correct.
func TestInviteSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	mgr := ws.NewRoomManager(log, dir)
	l := NewLobby(mgr)

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
	mgr2 := ws.NewRoomManager(log, dir)
	l2 := NewLobby(mgr2)
	l2.RestoreFromDisk(log)

	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatalf("restored lobby does not know the game: %v", err)
	}
	if back.InviteToken != meta.InviteToken {
		t.Errorf("invite token = %q, want %q — players cannot get back in", back.InviteToken, meta.InviteToken)
	}
	if back.SpectatorInvite != meta.SpectatorInvite {
		t.Errorf("spectator invite = %q, want %q", back.SpectatorInvite, meta.SpectatorInvite)
	}
	if back.Name != "Friday Night Commander" {
		t.Errorf("name = %q, want the original", back.Name)
	}
	if len(back.Players) != 2 {
		t.Fatalf("seat list has %d entries, want 2", len(back.Players))
	}
	if back.Players[0].Name != "Alice" || back.Players[1].Name != "Bob" {
		t.Errorf("seat list = %+v, want Alice then Bob", back.Players)
	}

	// And the invite still actually WORKS — the token round-tripped
	// as a string, but the check is constant-time against the stored
	// copy, so prove the whole path rather than the field.
	if _, _, err := l2.Join(back.ID, back.InviteToken, "Carol"); err != nil {
		t.Errorf("restored invite was rejected: %v", err)
	}
}

// TestGameWithoutMetadataIsDropped: a restored engine snapshot whose
// lobby metadata is missing is not a usable game — nobody can be
// invited to it — so it must not be left registered and half-visible.
func TestGameWithoutMetadataIsDropped(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	mgr := ws.NewRoomManager(log, dir)
	l := NewLobby(mgr)
	meta, err := l.Create("Doomed")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join: %v", err)
	}

	// Simulate the metadata being lost while the engine snapshot
	// survived.
	if err := os.Remove(metaPath(dir, meta.ID)); err != nil {
		t.Fatalf("remove meta: %v", err)
	}

	mgr2 := ws.NewRoomManager(log, dir)
	l2 := NewLobby(mgr2)
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

// TestDeletedGameStaysDeleted — the metadata must go with the game,
// or the next boot resurrects something the operator removed.
func TestDeletedGameStaysDeleted(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	mgr := ws.NewRoomManager(log, dir)
	l := NewLobby(mgr)
	meta, err := l.Create("Transient")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join: %v", err)
	}
	if err := l.Delete(meta.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	mgr2 := ws.NewRoomManager(log, dir)
	l2 := NewLobby(mgr2)
	l2.RestoreFromDisk(log)
	if _, err := l2.Get(meta.ID); err == nil {
		t.Error("a deleted game came back after a restart")
	}
}

// TestOrphanMetadataIsPruned keeps the data directory from growing a
// file per game forever.
func TestOrphanMetadataIsPruned(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	orphan := uuid.New()
	if err := writeFileAtomic0600(metaPath(dir, orphan), []byte(`{"name":"ghost"}`)); err != nil {
		t.Fatalf("seed orphan: %v", err)
	}

	mgr := ws.NewRoomManager(log, dir)
	l := NewLobby(mgr)
	l.RestoreFromDisk(log)

	if _, err := os.Stat(metaPath(dir, orphan)); !os.IsNotExist(err) {
		t.Errorf("orphan metadata survived the restore pass (err=%v)", err)
	}
}

// TestMetadataIsOwnerOnly — the file holds invite tokens, so its
// permissions are load-bearing.
func TestMetadataIsOwnerOnly(t *testing.T) {
	dir := t.TempDir()
	mgr := ws.NewRoomManager(quietLogger(), dir)
	l := NewLobby(mgr)
	meta, err := l.Create("Secrets")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	info, err := os.Stat(metaPath(dir, meta.ID))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("metadata mode = %04o, want 0600 — the file contains invite tokens", perm)
	}
}
