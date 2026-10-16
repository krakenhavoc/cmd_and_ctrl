package lobby

// restore_deferred_test.go covers the hot restore of tables the boot
// deferred (ADR 0148's 2026-10-13 amendment, #3103): they come back as
// load frees up, never over the cap, and their restore points and rows
// stay consistent. buildRestoreFixture (restore_capacity_test.go) is the
// process a crash leaves behind.

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// bootDeferring is the boot of the fixture under a cap of 12: the human
// table (4), the one-bot table D (4) and the archived table (free) come
// back; the four-bot tables A and C (10 each) are deferred.
func bootDeferring(t *testing.T) (restoreFixture, *Lobby) {
	t.Helper()
	f := buildRestoreFixture(t)
	l, _ := newDurableLobby(t, f.dir)
	l.SetCapacity(capLimits(12), nil)
	if n := l.RestoreFromDisk(quietLogger()); n != 3 {
		t.Fatalf("the boot restored %d tables, want 3", n)
	}
	return f, l
}

func fileBytes(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func TestHotRestoreBringsDeferredTablesBackAsLoadFreesAndNeverPassesTheCap(t *testing.T) {
	f, l := bootDeferring(t)
	aBefore := fileBytes(t, f.restorePoint(f.botsA))
	retiredRoom := l.RoomOf(f.retired)
	if retiredRoom == nil {
		t.Fatal("the archived table should be resident")
	}

	// Load is 8 of 12: neither four-bot table (10) fits. Nothing moves,
	// and the pass says so rather than failing.
	res := l.RestoreDeferred(quietLogger())
	if res.Restored != 0 || res.StillDeferred != 2 || res.Failed != 0 {
		t.Fatalf("with no room: %+v, want 0 restored and 2 still deferred", res)
	}
	if res.Load != 8 || res.LoadLimit != 12 {
		t.Errorf("load %d / %d, want 8 / 12", res.Load, res.LoadLimit)
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_deferred_games", nil); got != 2 {
		t.Errorf("cmdctrl_boot_restore_deferred_games = %v, want 2", got)
	}

	// Free the load: the two resident tables people were at go.
	for _, id := range []uuid.UUID{f.human, f.botsD} {
		if err := l.Delete(id); err != nil {
			t.Fatal(err)
		}
	}

	// Load 0 of 12: room for ONE four-bot table (10), and 20 would pass
	// the cap, so the second stays. C is newer than A, so C comes first.
	res = l.RestoreDeferred(quietLogger())
	if res.Restored != 1 || res.StillDeferred != 1 || res.Failed != 0 {
		t.Fatalf("with room for one: %+v, want 1 restored and 1 still deferred", res)
	}
	if res.Load != 10 || res.LoadLimit != 12 {
		t.Errorf("load %d / %d after the pass, want 10 / 12", res.Load, res.LoadLimit)
	}
	if got := l.CapacityUsage().Load; got > 12 {
		t.Errorf("the pass took the load to %d, over the cap of 12", got)
	}
	meta, err := l.Get(f.botsC)
	if err != nil {
		t.Fatalf("the newest deferred table did not come back: %v", err)
	}
	if len(meta.Players) != 4 || meta.State != "lobby" {
		t.Errorf("restored table C = %d seats, state %q; want 4 seats in lobby", len(meta.Players), meta.State)
	}
	if l.RoomOf(f.botsC) == nil {
		t.Error("table C has a lobby entry but no room")
	}
	if _, err := l.Get(f.botsA); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("table A should still be deferred, got err=%v", err)
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_deferred_games", nil); got != 1 {
		t.Errorf("cmdctrl_boot_restore_deferred_games = %v, want 1", got)
	}

	// The table still waiting is exactly as the crash left it: its
	// restore point byte for byte, its rows still saying lobby.
	if string(fileBytes(t, f.restorePoint(f.botsA))) != string(aBefore) {
		t.Error("a deferred restore point was modified by the hot pass")
	}
	ctx, cancel := storeCtx()
	rec, seats, err := l.store.LoadGame(ctx, f.botsA)
	cancel()
	if err != nil || rec.State != "lobby" || len(seats) != 4 {
		t.Errorf("deferred table A rows: err=%v state=%q seats=%d; want lobby with 4 seats", err, rec.State, len(seats))
	}
	// The restored table's rows agree with its room, and a table that was
	// already live was not rebuilt.
	ctx, cancel = storeCtx()
	rec, seats, err = l.store.LoadGame(ctx, f.botsC)
	cancel()
	if err != nil || rec.State != "lobby" || len(seats) != 4 {
		t.Errorf("restored table C rows: err=%v state=%q seats=%d; want lobby with 4 seats", err, rec.State, len(seats))
	}
	if l.RoomOf(f.retired) != retiredRoom {
		t.Error("the hot pass replaced a room that was already live")
	}

	// A second pass with the cap still full changes nothing.
	again := l.RestoreDeferred(quietLogger())
	if again.Restored != 0 || again.StillDeferred != 1 {
		t.Errorf("a second pass: %+v, want nothing restored and 1 deferred", again)
	}

	// Free C, and A takes its place.
	if err := l.Delete(f.botsC); err != nil {
		t.Fatal(err)
	}
	last := l.RestoreDeferred(quietLogger())
	if last.Restored != 1 || last.StillDeferred != 0 || last.Load != 10 {
		t.Errorf("the last pass: %+v, want A restored, none deferred, load 10", last)
	}
	if _, err := l.Get(f.botsA); err != nil {
		t.Errorf("table A did not come back: %v", err)
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_deferred_games", nil); got != 0 {
		t.Errorf("cmdctrl_boot_restore_deferred_games = %v with nothing waiting, want 0", got)
	}
}

// With the cap raised (or switched off) the same call restores every
// deferred table, as a restart would.
func TestHotRestoreAfterTheCapIsLiftedBringsEverythingBack(t *testing.T) {
	f, l := bootDeferring(t)
	l.SetCapacity(capLimits(0), nil)
	res := l.RestoreDeferred(quietLogger())
	if res.Restored != 2 || res.StillDeferred != 0 {
		t.Fatalf("with the cap off: %+v, want both deferred tables restored", res)
	}
	for name, id := range map[string]uuid.UUID{"A": f.botsA, "C": f.botsC} {
		if _, err := l.Get(id); err != nil {
			t.Errorf("table %s did not come back: %v", name, err)
		}
	}
}

// The reservation is what stops a create that lands mid-restore from
// overshooting: the admitted cost is held against the cap from the
// admission decision until the rooms are registered.
func TestHotRestoreHoldsItsAdmittedLoadAgainstConcurrentCreates(t *testing.T) {
	f, l := bootDeferring(t)
	for _, id := range []uuid.UUID{f.human, f.botsD} {
		if err := l.Delete(id); err != nil {
			t.Fatal(err)
		}
	}
	held := 0
	admit := l.admission(quietLogger(), &held)
	deferred := admit([]ws.RestoreCandidate{
		{GameID: f.botsC, Bots: 4},
		{GameID: f.botsA, Bots: 4},
	})
	if held != 10 || len(deferred) != 1 {
		t.Fatalf("admission held %d and deferred %d, want 10 held and 1 deferred", held, len(deferred))
	}
	// Cap 12, 10 held: a new empty table (2) fits, a bot-filled one does
	// not. With the hold missing a 10-cost create would pass here.
	if got := l.CapacityUsage().Load; got != 10 {
		t.Errorf("load while the restore is in flight = %d, want the 10 it holds", got)
	}
	release, _, err := l.reserveLoad(capacity.Defaults().TableCost(0, 4), "table_load")
	if err == nil {
		release()
		t.Fatal("a create passed while the restore held the room it was admitted into")
	}
	if !errors.Is(err, capacity.ErrServerFull) {
		t.Errorf("err = %v, want ErrServerFull", err)
	}
	l.releaseHeld(held)
	if got := l.CapacityUsage().Load; got != 0 {
		t.Errorf("load after the hold is released = %d, want 0", got)
	}
}

// The same call from a lobby with no durable store (no restore points
// to read) is a clean zero, not an error.
func TestHotRestoreWithNothingOnDiskIsANoOp(t *testing.T) {
	l := NewLobby(ws.NewRoomManager(quietLogger(), ""))
	if res := l.RestoreDeferred(quietLogger()); res.Restored != 0 || res.StillDeferred != 0 || res.Failed != 0 {
		t.Errorf("no store: %+v, want zeros", res)
	}
}

func TestAdminRestoreDeferredRouteIsAdminOnlyAndReportsTheCounts(t *testing.T) {
	f, l := bootDeferring(t)
	for _, id := range []uuid.UUID{f.human, f.botsD} {
		if err := l.Delete(id); err != nil {
			t.Fatal(err)
		}
	}
	a := newHMAC(t)
	srv := httptest.NewServer(Handler(Config{Lobby: l, Auth: a, AdminToken: "shared-admin-token"}))
	t.Cleanup(srv.Close)

	// Not an admin: refused, and nothing came back.
	resp := postJSON(t, srv, "/admin/capacity/restore-deferred", "", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("without a credential: %d, want 401", resp.StatusCode)
	}
	if _, err := l.Get(f.botsC); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("a refused call restored a table (err=%v)", err)
	}

	resp = postJSON(t, srv, "/admin/capacity/restore-deferred", adminSession(t, a), nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("as an admin: %d, want 200", resp.StatusCode)
	}
	var got RestoreDeferredResult
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	want := RestoreDeferredResult{Restored: 1, StillDeferred: 1, Load: 10, LoadLimit: 12}
	if got != want {
		t.Errorf("response = %+v, want %+v", got, want)
	}
}
