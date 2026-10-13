package ws

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// RestoreRoomsAdmitting (ADR 0148, #2977): the admission step between
// reading the restore points and rebuilding them. A deferred game's
// file is not opened for restore, modified or removed; an unreadable
// file is still reported as it always was and never offered to admit.

// persisted makes a game in mgr and plays one move so its restore point
// reaches disk.
func persisted(t *testing.T, mgr *RoomManager) *game.Game {
	t.Helper()
	g := newPersistGame(t)
	room := mgr.Create(g)
	seat0 := g.Seats[0].ID
	if _, _, err := room.Apply(seat0, func() error { return g.DrawCard(seat0) }); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return g
}

func TestRestoreAdmissionDefersWithoutTouchingTheFile(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)
	a, b := persisted(t, mgr), persisted(t, mgr)

	pathB := filepath.Join(dir, "restore", b.ID.String()+".json")
	before, err := os.ReadFile(pathB)
	if err != nil {
		t.Fatalf("no restore point for the second game: %v", err)
	}

	var offered []RestoreCandidate
	mgr2 := restoreTestManager(t, dir)
	outcomes := mgr2.RestoreRoomsAdmitting(func(cands []RestoreCandidate) map[uuid.UUID]bool {
		offered = cands
		return map[uuid.UUID]bool{b.ID: true}
	})

	if len(offered) != 2 {
		t.Fatalf("admit was offered %d candidates, want 2", len(offered))
	}
	for _, c := range offered {
		// newPersistGame seats two humans.
		if c.Humans != 2 || c.Bots != 0 || c.TakenAt.IsZero() {
			t.Errorf("candidate %+v: want 2 humans, 0 bots and a capture time", c)
		}
	}
	var restored, deferred int
	for _, o := range outcomes {
		switch {
		case o.Restored():
			restored++
			if o.GameID != a.ID {
				t.Errorf("restored %s, want the admitted game", o.GameID)
			}
		case o.Deferred:
			deferred++
			if o.GameID != b.ID || o.Reason != ReasonOverCapacity || o.Err != nil {
				t.Errorf("deferred outcome %+v", o)
			}
		}
	}
	if restored != 1 || deferred != 1 {
		t.Fatalf("restored %d, deferred %d; want 1 and 1", restored, deferred)
	}
	if mgr2.Get(b.ID) != nil {
		t.Error("a deferred game was registered")
	}
	after, err := os.ReadFile(pathB)
	if err != nil {
		t.Fatalf("the deferred restore point is gone: %v", err)
	}
	if string(after) != string(before) {
		t.Error("the deferred restore point was modified")
	}

	LogRestoreSummary(restoreTestManager(t, t.TempDir()).log, outcomes)
	if got := metricstest.Value(t, "cmdctrl_boot_restore_deferred_games", nil); got != 1 {
		t.Errorf("cmdctrl_boot_restore_deferred_games = %v, want 1", got)
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_games", metricstest.L{"outcome": "abandoned"}); got != 0 {
		t.Errorf("a deferred game must not count as abandoned: %v", got)
	}
}

func TestRestoreAdmissionIsNeverAskedAboutAnUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)
	persisted(t, mgr)
	if err := os.WriteFile(filepath.Join(dir, "restore", uuid.NewString()+".json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	asked := 0
	outcomes := restoreTestManager(t, dir).RestoreRoomsAdmitting(func(cands []RestoreCandidate) map[uuid.UUID]bool {
		asked = len(cands)
		return nil
	})
	if asked != 1 {
		t.Errorf("admit saw %d candidates, want only the readable one", asked)
	}
	var bad int
	for _, o := range outcomes {
		if o.Reason == ReasonDecodeError {
			bad++
		}
	}
	if len(outcomes) != 2 || bad != 1 {
		t.Errorf("outcomes %d (decode errors %d), want the good game restored and the bad file reported", len(outcomes), bad)
	}
}

func TestRestoreRoomsIsRestoreRoomsAdmittingEverything(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)
	persisted(t, mgr)
	outcomes := restoreTestManager(t, dir).RestoreRoomsAdmitting(nil)
	if len(outcomes) != 1 || !outcomes[0].Restored() {
		t.Fatalf("a nil admission must restore everything: %+v", outcomes)
	}
}
