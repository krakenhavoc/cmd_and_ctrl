package lobby

// restore_capacity_test.go covers the boot restore pass under the load
// cap (ADR 0148, #2977): which tables come back, and that the rest are
// left exactly as they were for a boot that has room.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// restoreFixture builds the process a crash leaves behind, in this
// creation order (oldest first), each table written to disk before the
// next is made so the restore points' capture times are ordered:
//
//	human   2 human seats                    cost 4   (oldest)
//	botsA   4 bots                           cost 10
//	botsD   1 bot                            cost 4
//	botsC   4 bots                           cost 10  (newest)
//	retired 4 bots, then archived            cost 0
type restoreFixture struct {
	dir                                 string
	human, botsA, botsD, botsC, retired uuid.UUID
}

func buildRestoreFixture(t *testing.T) restoreFixture {
	t.Helper()
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	f := restoreFixture{dir: dir}
	tick := func() { time.Sleep(5 * time.Millisecond) }

	h, err := l.Create("human")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"Alice", "Bob"} {
		if _, _, err := l.Join(h.ID, h.InviteToken, n); err != nil {
			t.Fatal(err)
		}
	}
	f.human = h.ID
	tick()
	mk := func(name string, bots int) uuid.UUID {
		m, err := l.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		fillBots(t, l, m.ID, bots)
		tick()
		return m.ID
	}
	f.botsA = mk("bots A", 4)
	f.botsD = mk("bots D", 1)
	f.botsC = mk("bots C", 4)
	f.retired = mk("retired", 4)
	if _, err := l.SetArchived(f.retired, true); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f restoreFixture) restorePoint(id uuid.UUID) string {
	return filepath.Join(f.dir, "restore", id.String()+".json")
}

func TestRestorePassKeepsTheLoadUnderTheCapAndLeavesTheRestOnDisk(t *testing.T) {
	f := buildRestoreFixture(t)

	// What the files hold now, to prove a deferred one is untouched.
	before := map[uuid.UUID][]byte{}
	for _, id := range []uuid.UUID{f.botsA, f.botsC} {
		raw, err := os.ReadFile(f.restorePoint(id))
		if err != nil {
			t.Fatalf("the fixture wrote no restore point for %s: %v", id, err)
		}
		before[id] = raw
	}

	// The boot, with a cap of 12: the human table first (4), then the
	// bot tables newest first: C (10) does not fit and is deferred, D
	// (4) does, A (10) does not. The archived table costs nothing.
	l2, _ := newDurableLobby(t, f.dir)
	l2.SetCapacity(capLimits(12), nil)
	if n := l2.RestoreFromDisk(quietLogger()); n != 3 {
		t.Fatalf("restored %d tables, want 3 (human, D, and the archived one)", n)
	}
	for name, id := range map[string]uuid.UUID{"human": f.human, "bots D": f.botsD, "archived": f.retired} {
		if _, err := l2.Get(id); err != nil {
			t.Errorf("%s should have come back: %v", name, err)
		}
	}
	for name, id := range map[string]uuid.UUID{"bots A": f.botsA, "bots C": f.botsC} {
		if _, err := l2.Get(id); !errors.Is(err, ErrGameNotFound) {
			t.Errorf("%s should have been deferred, got err=%v", name, err)
		}
	}
	if got := l2.CapacityUsage().Load; got != 4+4 {
		t.Errorf("load after the restore pass = %d, want 8 (human 4 + D 4; archived free)", got)
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_deferred_games", nil); got != 2 {
		t.Errorf("cmdctrl_boot_restore_deferred_games = %v, want 2", got)
	}

	// The deferred tables are intact: restore point byte for byte, the
	// games row still saying lobby (NOT ended), seats row untouched.
	for _, id := range []uuid.UUID{f.botsA, f.botsC} {
		raw, err := os.ReadFile(f.restorePoint(id))
		if err != nil {
			t.Fatalf("a deferred restore point is gone: %v", err)
		}
		if string(raw) != string(before[id]) {
			t.Errorf("a deferred restore point was modified")
		}
		ctx, cancel := storeCtx()
		rec, seats, err := l2.store.LoadGame(ctx, id)
		cancel()
		if err != nil {
			t.Fatalf("a deferred table's rows are gone: %v", err)
		}
		if rec.State != "lobby" {
			t.Errorf("a deferred table's row says %q; it must be left as it was (lobby), not ended", rec.State)
		}
		if len(seats) != 4 {
			t.Errorf("a deferred table lost seat rows: %d, want 4", len(seats))
		}
	}

	// A later boot with room brings every one of them back.
	l3, _ := newDurableLobby(t, f.dir)
	l3.SetCapacity(capLimits(1000), nil)
	if n := l3.RestoreFromDisk(quietLogger()); n != 5 {
		t.Fatalf("a boot with room restored %d tables, want all 5", n)
	}
	for _, id := range []uuid.UUID{f.botsA, f.botsC} {
		if _, err := l3.Get(id); err != nil {
			t.Errorf("a deferred table did not come back on the next boot: %v", err)
		}
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_deferred_games", nil); got != 0 {
		t.Errorf("cmdctrl_boot_restore_deferred_games = %v after a boot with room, want 0", got)
	}
}

func TestRestorePassPutsTablesWithPeopleFirstEvenWhenOlder(t *testing.T) {
	f := buildRestoreFixture(t)
	// Room for the human table and nothing else but D: the human table
	// is the OLDEST restore point, and still comes first.
	l, _ := newDurableLobby(t, f.dir)
	l.SetCapacity(capLimits(4), nil)
	l.RestoreFromDisk(quietLogger())
	if _, err := l.Get(f.human); err != nil {
		t.Fatalf("the table with people at it must come back first: %v", err)
	}
	for name, id := range map[string]uuid.UUID{"A": f.botsA, "C": f.botsC, "D": f.botsD} {
		if _, err := l.Get(id); !errors.Is(err, ErrGameNotFound) {
			t.Errorf("bot table %s should have been deferred (err=%v)", name, err)
		}
	}
}

func TestRestorePassWithTheCapOffRestoresEverything(t *testing.T) {
	f := buildRestoreFixture(t)
	l, _ := newDurableLobby(t, f.dir)
	// SetCapacity never called, or MaxLoad 0: the cap is off.
	if n := l.RestoreFromDisk(quietLogger()); n != 5 {
		t.Fatalf("restored %d, want 5", n)
	}
	l2, _ := newDurableLobby(t, f.dir)
	l2.SetCapacity(capLimits(0), nil)
	if n := l2.RestoreFromDisk(quietLogger()); n != 5 {
		t.Fatalf("with MaxLoad 0: restored %d, want 5", n)
	}
}

// An admin deleting a table the boot deferred removes its restore point
// and rows, as for any unresident table: the deferral parks a game, it
// does not make it undeletable.
func TestADeferredTableCanStillBeDeleted(t *testing.T) {
	f := buildRestoreFixture(t)
	l, _ := newDurableLobby(t, f.dir)
	l.SetCapacity(capLimits(4), nil)
	l.RestoreFromDisk(quietLogger())
	if err := l.Delete(f.botsC); err != nil {
		t.Fatalf("Delete of a deferred table: %v", err)
	}
	if _, err := os.Stat(f.restorePoint(f.botsC)); !os.IsNotExist(err) {
		t.Errorf("the deleted table's restore point is still on disk: %v", err)
	}
}
