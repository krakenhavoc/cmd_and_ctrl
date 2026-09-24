package ws

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// persist_unknown_effect_test.go is ADR 0041 phase 3's P4 at the boot
// path (#1497): a restore point naming an effect this binary cannot
// interpret — which only a NEWER binary can have written, so this is a
// rollback — abandons the game with its own reason and KEEPS the file,
// so rolling forward again brings the table back.
func TestARestorePointWithAnUnknownEffectIsKeptForRollForward(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)
	g := newPersistGame(t)
	room := mgr.Create(g)
	bear := uuid.New()
	if _, _, err := room.Apply(g.Seats[0].ID, func() error {
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
				Power: 2, Toughness: 2, Owner: g.Seats[0].ID, Controller: g.Seats[0].ID,
			})
			g.RegisterScopedEffectForEffect(uuid.Nil,
				[]game.AffectedObject{{ID: bear}},
				[]game.Mod{game.AddSubtypesMod("Orc")}, game.IndefiniteDuration(), "test")
		})
		return nil
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	path := restorePointPath(dir, g.ID)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the scoped effect did not leave a restore point: %v", err)
	}
	// The file a later build would write: a kind this one has never
	// heard of.
	future := bytes.Replace(raw, []byte(`"addSubtypes"`), []byte(`"grantAbilitiesFromTheFuture"`), 1)
	if bytes.Equal(future, raw) {
		t.Fatal("setup: the restore point does not name the mod kind")
	}
	if err := os.WriteFile(path, future, 0o600); err != nil {
		t.Fatal(err)
	}

	mgr = NewRoomManager(slog.New(slog.NewTextHandler(io.Discard, nil)), dir)
	outcomes := mgr.RestoreRooms()
	if len(outcomes) != 1 {
		t.Fatalf("outcomes = %+v, want one", outcomes)
	}
	if outcomes[0].Restored() {
		t.Fatal("a restore point naming an unknown effect was restored")
	}
	if outcomes[0].Reason != ReasonUnknownEffectKey {
		t.Errorf("reason = %q, want %q", outcomes[0].Reason, ReasonUnknownEffectKey)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the restore point was not kept for the roll-forward: %v", err)
	}
}
