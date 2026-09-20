package lobby

// rotate_invite_test.go covers issue #1038: revoke-and-re-mint for a
// lost invite link. The scenario that makes the feature worth having
// is TestRotateInviteWorksWithoutTheOldPlaintext — a second process
// that never saw the first process's plaintext token must still be
// able to rotate, because Store.RotateInvite revokes by (game, kind),
// not by hash.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRotateInviteHappyPath(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Rotate me")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	oldToken := meta.InviteToken

	newToken, updated, err := l.RotateInvite(meta.ID, InvitePlayer)
	if err != nil {
		t.Fatalf("RotateInvite: %v", err)
	}
	if newToken == "" || newToken == oldToken {
		t.Fatalf("RotateInvite returned an unusable token: %q (old %q)", newToken, oldToken)
	}
	if updated.InviteToken != newToken {
		t.Errorf("returned meta.InviteToken = %q, want %q", updated.InviteToken, newToken)
	}

	// Old token is refused.
	if _, _, err := l.Join(meta.ID, oldToken, "Alice"); err != ErrInvalidInvite {
		t.Errorf("join with old token: got %v, want ErrInvalidInvite", err)
	}
	// New token is accepted.
	if _, _, err := l.Join(meta.ID, newToken, "Alice"); err != nil {
		t.Errorf("join with new token: %v", err)
	}
	// GET-equivalent (Get) reflects the new plaintext.
	got, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.InviteToken != newToken {
		t.Errorf("Get().InviteToken = %q, want %q", got.InviteToken, newToken)
	}
}

func TestRotateInviteLeavesOtherKindWorking(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Rotate one kind")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	specToken := meta.SpectatorInvite

	if _, _, err := l.RotateInvite(meta.ID, InvitePlayer); err != nil {
		t.Fatalf("RotateInvite(player): %v", err)
	}

	// The untouched spectator invite still works.
	if _, err := l.Spectate(meta.ID, specToken); err != nil {
		t.Errorf("spectate with untouched invite: %v", err)
	}
}

func TestRotateSpectatorInviteLeavesPlayerInviteWorking(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Rotate spectator")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	playerToken := meta.InviteToken
	oldSpecToken := meta.SpectatorInvite

	newSpecToken, _, err := l.RotateInvite(meta.ID, InviteSpectator)
	if err != nil {
		t.Fatalf("RotateInvite(spectator): %v", err)
	}

	if _, err := l.Spectate(meta.ID, oldSpecToken); err != ErrInvalidInvite {
		t.Errorf("spectate with old spectator token: got %v, want ErrInvalidInvite", err)
	}
	if _, err := l.Spectate(meta.ID, newSpecToken); err != nil {
		t.Errorf("spectate with new spectator token: %v", err)
	}
	if _, _, err := l.Join(meta.ID, playerToken, "Alice"); err != nil {
		t.Errorf("join with untouched player invite: %v", err)
	}
}

func TestRotateInviteRejectsUnknownKind(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Bad kind")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, _, err := l.RotateInvite(meta.ID, InviteKind("banana")); err != ErrInvalidInviteKind {
		t.Errorf("RotateInvite bad kind: got %v, want ErrInvalidInviteKind", err)
	}
}

func TestRotateInviteUnknownGame(t *testing.T) {
	l := newTestLobby(t)
	if _, _, err := l.RotateInvite(uuid.New(), InvitePlayer); err != ErrGameNotFound {
		t.Errorf("RotateInvite unknown game: got %v, want ErrGameNotFound", err)
	}
}

// TestRotateInviteWorksWithoutTheOldPlaintext is the scenario #1038
// exists for: a game created by one process, and rotated by a SECOND
// process (a fresh boot against the same database) that never held
// the first process's plaintext invite token — the exact situation a
// restart leaves the admin in. Store.RotateInvite revokes by
// (game_id, kind), not by hash, so it needs nothing from the token it
// is replacing.
func TestRotateInviteWorksWithoutTheOldPlaintext(t *testing.T) {
	dir := t.TempDir()

	l1, _ := newDurableLobby(t, dir)
	meta, err := l1.Create("Restart me")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	oldToken := meta.InviteToken
	// A restore point is only written once the room has taken a
	// mutation; a bare Create leaves nothing on disk for
	// RestoreRooms to find.
	if _, _, err := l1.Join(meta.ID, oldToken, "Zed"); err != nil {
		t.Fatalf("Join: %v", err)
	}

	// A second process opens the same database. It never saw oldToken
	// in memory — only its hash reached the store.
	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("RestoreFromDisk restored %d games, want 1", n)
	}
	// Confirm the fresh process really doesn't have the plaintext.
	restored, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get on restored lobby: %v", err)
	}
	if restored.InviteToken != "" {
		t.Fatalf("restored entry unexpectedly carries a plaintext invite: %q", restored.InviteToken)
	}

	newToken, _, err := l2.RotateInvite(meta.ID, InvitePlayer)
	if err != nil {
		t.Fatalf("RotateInvite on the restored process: %v", err)
	}
	if newToken == "" {
		t.Fatal("RotateInvite returned an empty token")
	}

	// The pre-restart token is dead everywhere...
	if _, _, err := l2.Join(meta.ID, oldToken, "Alice"); err != ErrInvalidInvite {
		t.Errorf("join with pre-restart token: got %v, want ErrInvalidInvite", err)
	}
	// ...and the freshly rotated one works, on the process that
	// rotated it...
	if _, _, err := l2.Join(meta.ID, newToken, "Alice"); err != nil {
		t.Errorf("join with rotated token on l2: %v", err)
	}

	// ...and would work on a third process too, since resolveInvite
	// only ever checks the store.
	l3, _ := newDurableLobby(t, dir)
	if n := l3.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("l3 RestoreFromDisk restored %d games, want 1", n)
	}
	if _, _, err := l3.Join(meta.ID, newToken, "Bob"); err != nil {
		t.Errorf("join with rotated token on l3: %v", err)
	}
}

// TestRotateInviteConcurrent hammers RotateInvite for the same game
// and kind from many goroutines. Every call must succeed (no store
// error, no duplicate-hash collision under crypto/rand), and exactly
// one token must be usable afterwards — l.mu serializes the whole
// operation, so this also stands in as the "concurrent rotation is
// handled sanely" test. Run with -race.
func TestRotateInviteConcurrent(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Concurrent rotate")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	const n = 20
	tokens := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			tok, _, err := l.RotateInvite(meta.ID, InvitePlayer)
			tokens[i] = tok
			errs[i] = err
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("rotation %d failed: %v", i, err)
		}
		if tokens[i] == "" {
			t.Fatalf("rotation %d returned an empty token", i)
		}
		if seen[tokens[i]] {
			t.Fatalf("rotation %d reused a token another rotation already returned", i)
		}
		seen[tokens[i]] = true
	}

	// Every token but the very last write is now revoked; exactly one
	// resolves. We don't know which one l.mu let through last, so
	// just assert exactly one of the n tokens still joins.
	working := 0
	var workingToken string
	for tok := range seen {
		if _, _, err := l.Preview(meta.ID, tok); err == nil {
			working++
			workingToken = tok
		}
	}
	if working != 1 {
		t.Errorf("%d of %d rotated tokens still resolve, want exactly 1", working, n)
	}
	if _, _, err := l.Join(meta.ID, workingToken, "Winner"); err != nil {
		t.Errorf("join with the surviving token: %v", err)
	}
}

// TestRotateInviteStoreFailureLeavesOldTokenAlone: if the store write
// fails, the in-memory plaintext must NOT change — resolveInvite only
// ever checks the store, so a token that never reached it can never
// be redeemed, and silently swapping the displayed token for one that
// doesn't work would be worse than doing nothing.
func TestRotateInviteStoreFailureLeavesOldTokenAlone(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("Store failure")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	oldToken := meta.InviteToken

	l.store = failingRotateStore{Store: l.store}

	if _, _, err := l.RotateInvite(meta.ID, InvitePlayer); err == nil {
		t.Fatal("RotateInvite over a failing store returned no error")
	}

	got, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.InviteToken != oldToken {
		t.Errorf("InviteToken changed despite a failed store write: got %q, want %q", got.InviteToken, oldToken)
	}
	if _, _, err := l.Join(meta.ID, oldToken, "Alice"); err != nil {
		t.Errorf("the pre-rotation token stopped working: %v", err)
	}
}

// failingRotateStore wraps a Store and fails every RotateInvite call,
// to exercise Lobby.RotateInvite's error path without a real store
// outage.
type failingRotateStore struct {
	Store
}

func (failingRotateStore) RotateInvite(context.Context, uuid.UUID, InviteKind, InviteRecord, time.Time) error {
	return errors.New("rotate_invite_test: simulated store failure")
}
