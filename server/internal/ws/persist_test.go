package ws

import (
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// persist_test.go is the end-to-end proof of the thing the owner
// actually asked for: deploy new code without destroying live games.
//
// The simulation is deliberately blunt — build a room, play into it,
// throw the whole RoomManager away, build a new one over the same data
// directory, and check the table is still there. That is what a deploy
// does.

func restoreTestManager(t *testing.T, dir string) *RoomManager {
	t.Helper()
	return NewRoomManager(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})), dir)
}

// newPersistGame builds a started 2-player game with a persistable
// random source, so the whole board is a valid restore point.
func newPersistGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 0, 12)
		deck = append(deck, game.Card{
			InstanceID:  uuid.New(),
			Name:        "Test Commander",
			OracleID:    "oracle-commander",
			TypeLine:    "Legendary Creature — Test",
			IsCommander: true,
		})
		for j := 0; j < 11; j++ {
			deck = append(deck, game.Card{
				InstanceID: uuid.New(),
				Name:       "Forest",
				OracleID:   "oracle-forest",
				TypeLine:   "Basic Land — Forest",
			})
		}
		if _, err := g.AddPlayer("P", deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(rand.NewPCG(3, 5)); err != nil {
		t.Fatalf("StartWithSource: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// TestRestartPreservesALiveGame is the headline. A game played into,
// then subjected to a full process-equivalent restart, comes back with
// its board, its life totals and its sequence number.
func TestRestartPreservesALiveGame(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	g := newPersistGame(t)
	room := mgr.Create(g)
	seat0 := g.Seats[0].ID

	// Play. Each Apply is a chance for the restore point to be
	// written.
	for i := 0; i < 3; i++ {
		if _, _, err := room.Apply(seat0, func() error {
			return g.DrawCard(seat0)
		}); err != nil {
			t.Fatalf("Apply draw %d: %v", i, err)
		}
	}
	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 33 })
		return nil
	}); err != nil {
		t.Fatalf("Apply life change: %v", err)
	}

	wantSeq := room.Seq()
	wantHand := g.Seats[0].Hand.Size()
	wantLib := g.Seats[0].Library.Size()
	gameID := g.ID

	// --- the deploy: everything in memory goes away ---------------
	mgr = restoreTestManager(t, dir)
	outcomes := mgr.RestoreRooms()

	if len(outcomes) != 1 {
		t.Fatalf("restore produced %d outcomes, want 1", len(outcomes))
	}
	if !outcomes[0].Restored() {
		t.Fatalf("game was not restored: err=%v skipped=%q", outcomes[0].Err, outcomes[0].Skipped)
	}

	back := mgr.Get(gameID)
	if back == nil {
		t.Fatal("restored room is not registered with the manager")
	}
	if got := back.Seq(); got != wantSeq {
		t.Errorf("seq = %d, want %d — a restored room that restarts its counter hands reconnecting clients a seq they have already seen", got, wantSeq)
	}
	if got := back.Game.Seats[0].Life; got != 33 {
		t.Errorf("life = %d, want 33", got)
	}
	if got := back.Game.Seats[0].Hand.Size(); got != wantHand {
		t.Errorf("hand size = %d, want %d", got, wantHand)
	}
	if got := back.Game.Seats[0].Library.Size(); got != wantLib {
		t.Errorf("library size = %d, want %d", got, wantLib)
	}
	if got := back.Game.CurrentState(); got != game.StateActive {
		t.Errorf("state = %q, want active", got)
	}

	// The restored room is a working room: it accepts an action and
	// keeps advancing the same sequence.
	if _, seq, err := back.Apply(back.Game.Seats[0].ID, func() error {
		return back.Game.DrawCard(back.Game.Seats[0].ID)
	}); err != nil {
		t.Errorf("restored room rejected an action: %v", err)
	} else if seq != wantSeq+1 {
		t.Errorf("post-restore seq = %d, want %d", seq, wantSeq+1)
	}
}

// TestRestorePointHoldsAtTheLastCleanBoundary is the declared
// simplification, proved rather than promised. A game that walks into
// a state holding a live continuation stops updating its restore
// point, so a restart rewinds to the last state that can be rebuilt
// exactly — it does not resurrect the game wrong.
func TestRestorePointHoldsAtTheLastCleanBoundary(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	g := newPersistGame(t)
	room := mgr.Create(g)
	seat0 := g.Seats[0].ID

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 30 })
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	cleanSeq := room.Seq()

	// Now walk into a continuation: an ability on the stack whose
	// resolution behaviour is a Go closure.
	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() {
			g.Seats[0].Life = 20
			id := uuid.New()
			g.StackMeta = map[uuid.UUID]*game.StackItem{id: {
				ID:         id,
				Kind:       game.StackItemActivated,
				Controller: seat0,
				Label:      "some ability",
				Effect:     func(*game.Game, *game.StackItem) error { return nil },
			}}
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if room.Seq() == cleanSeq {
		t.Fatal("setup: the second apply did not advance seq")
	}

	// The restore point must still be the clean one.
	mgr = restoreTestManager(t, dir)
	outcomes := mgr.RestoreRooms()
	if len(outcomes) != 1 || !outcomes[0].Restored() {
		t.Fatalf("restore failed: %+v", outcomes)
	}
	back := mgr.Get(g.ID)
	if got := back.Seq(); got != cleanSeq {
		t.Errorf("restored seq = %d, want %d (the last continuation-free state)", got, cleanSeq)
	}
	if got := back.Game.Seats[0].Life; got != 30 {
		t.Errorf("life = %d, want 30 — the restore rewound past the clean boundary or did not rewind to it", got)
	}
	if n := len(back.Game.StackMeta); n != 0 {
		t.Errorf("restored game has %d stack items, want 0 — it should have rewound past the ability", n)
	}
}

// TestEndedGameIsNotResurrected — a finished table has nothing to
// resume, and leaving its file behind would have every future boot
// rebuild it.
func TestEndedGameIsNotResurrected(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	g := newPersistGame(t)
	room := mgr.Create(g)
	seat0 := g.Seats[0].ID

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 5 })
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(restorePointPath(dir, g.ID)); err != nil {
		t.Fatalf("setup: no restore point was written: %v", err)
	}

	if _, _, err := room.Apply(seat0, func() error {
		g.WithWriteLock(func() { g.State = game.StateEnded })
		return nil
	}); err != nil {
		t.Fatalf("Apply end: %v", err)
	}

	if _, err := os.Stat(restorePointPath(dir, g.ID)); !os.IsNotExist(err) {
		t.Errorf("restore point survived the game ending (err=%v)", err)
	}
	mgr = restoreTestManager(t, dir)
	if n := len(mgr.RestoreRooms()); n != 0 {
		t.Errorf("restore pass returned %d outcomes for an ended game, want 0", n)
	}
}

// TestFutureSchemaIsAbandonedNotGuessed is the version-skew policy at
// the boot layer: a file from a newer server is refused, the boot
// still succeeds, and the file is KEPT so rolling the binary forward
// brings the game back.
func TestFutureSchemaIsAbandonedNotGuessed(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	g := newPersistGame(t)
	room := mgr.Create(g)
	if _, _, err := room.Apply(g.Seats[0].ID, func() error { return nil }); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	path := restorePointPath(dir, g.ID)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read restore point: %v", err)
	}
	var file restorePointFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode: %v", err)
	}
	file.Snapshot.Schema = game.SnapshotSchemaVersion + 1
	bumped, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, bumped, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	mgr = restoreTestManager(t, dir)
	outcomes := mgr.RestoreRooms()
	if len(outcomes) != 1 {
		t.Fatalf("got %d outcomes, want 1", len(outcomes))
	}
	if outcomes[0].Restored() {
		t.Error("a snapshot from a newer server was restored; it must be abandoned instead")
	}
	if outcomes[0].Err == nil {
		t.Error("abandonment was not reported as an error")
	}
	if mgr.Get(g.ID) != nil {
		t.Error("abandoned game was registered anyway")
	}
	// The file must survive so a roll-forward recovers the game.
	if _, err := os.Stat(path); err != nil {
		t.Errorf("restore point was deleted; a roll-forward can no longer recover the game: %v", err)
	}
}

// TestCorruptRestorePointDoesNotStopTheBoot — one bad file must cost
// one game, not the server.
func TestCorruptRestorePointDoesNotStopTheBoot(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	good := newPersistGame(t)
	room := mgr.Create(good)
	if _, _, err := room.Apply(good.Seats[0].ID, func() error { return nil }); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// A second file that is not valid JSON.
	bad := uuid.New()
	if err := os.MkdirAll(filepath.Join(dir, "restore"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(restorePointPath(dir, bad), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	mgr = restoreTestManager(t, dir)
	outcomes := mgr.RestoreRooms()
	if len(outcomes) != 2 {
		t.Fatalf("got %d outcomes, want 2", len(outcomes))
	}
	if mgr.Get(good.ID) == nil {
		t.Error("the healthy game did not come back alongside the corrupt one")
	}
}

// TestDeleteRemovesTheRestorePoint — an operator who deletes a game
// must not see it return at the next deploy.
func TestDeleteRemovesTheRestorePoint(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	g := newPersistGame(t)
	room := mgr.Create(g)
	if _, _, err := room.Apply(g.Seats[0].ID, func() error { return nil }); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(restorePointPath(dir, g.ID)); err != nil {
		t.Fatalf("setup: no restore point: %v", err)
	}

	mgr.Delete(g.ID)

	if _, err := os.Stat(restorePointPath(dir, g.ID)); !os.IsNotExist(err) {
		t.Errorf("restore point survived Delete (err=%v)", err)
	}
}

// TestPersistenceDisabledWritesNothing — CMDCTRL_DATA_DIR="" is an
// explicit "no disk", and it must stay that way.
func TestPersistenceDisabledWritesNothing(t *testing.T) {
	mgr := restoreTestManager(t, "")
	g := newPersistGame(t)
	room := mgr.Create(g)
	if _, _, err := room.Apply(g.Seats[0].ID, func() error { return nil }); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if n := len(mgr.RestoreRooms()); n != 0 {
		t.Errorf("RestoreRooms returned %d outcomes with persistence disabled", n)
	}
}
