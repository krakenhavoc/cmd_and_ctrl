package ws

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics/metricstest"
)

// action_metrics_test.go is ADR 0123 §11's check on the engine half of
// §3: each commit path moves cmdctrl_actions_total and
// cmdctrl_action_apply_seconds by exactly one, with the labels
// action_metrics.go documents; the restore-point gauges read a built
// room; the boot gauges follow the restore pass.

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// seatKindsRoom is a lobby-state table with a person, a bot and an
// agent, and no disk.
func seatKindsRoom(t *testing.T) (r *Room, human, bot, agent uuid.UUID) {
	t.Helper()
	g := game.NewGame()
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		p, err := g.AddPlayer("P", []game.Card{{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"}})
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		ids = append(ids, p.ID)
	}
	if err := g.SetBot(ids[1], "heuristic", ""); err != nil {
		t.Fatalf("SetBot: %v", err)
	}
	if err := g.SetAgent(ids[2], "claude-code"); err != nil {
		t.Fatalf("SetAgent: %v", err)
	}
	return NewRoom(g, quietLog(), ""), ids[0], ids[1], ids[2]
}

func actionsTotal(t *testing.T, labels metricstest.L) float64 {
	t.Helper()
	return metricstest.Value(t, "cmdctrl_actions_total", labels)
}

func applySeconds(t *testing.T, result string) float64 {
	t.Helper()
	return metricstest.Value(t, "cmdctrl_action_apply_seconds", metricstest.L{"result": result})
}

var errRefused = errors.New("refused")

func ok() error     { return nil }
func refuse() error { return errRefused }

// Each path, applied and refused, moves exactly one series by one and
// observes exactly one lock hold, and the series is the one the
// labelling rule names.
func TestEachApplyPathCountsOnce(t *testing.T) {
	room, human, bot, agent := seatKindsRoom(t)

	cases := []struct {
		name   string
		commit func() error
		want   metricstest.L
	}{
		{"Apply", func() error { _, _, err := room.Apply(human, ok); return err },
			metricstest.L{"type": "other", "seat_kind": "human", "result": "applied"}},
		{"Apply refused", func() error { _, _, err := room.Apply(human, refuse); return err },
			metricstest.L{"type": "other", "seat_kind": "human", "result": "rejected"}},
		{"ApplyAction human", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: actions.TypePassPriority, Actor: human}, ok)
			return err
		}, metricstest.L{"type": "pass_priority", "seat_kind": "human", "result": "applied"}},
		{"ApplyAction refused", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: actions.TypeCastSpell, Actor: human}, refuse)
			return err
		}, metricstest.L{"type": "cast_spell", "seat_kind": "human", "result": "rejected"}},
		{"ApplyAction bot", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: actions.TypePlayCard, Actor: bot}, ok)
			return err
		}, metricstest.L{"type": "play_card", "seat_kind": "bot", "result": "applied"}},
		{"ApplyAction agent", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: actions.TypeTap, Actor: agent}, ok)
			return err
		}, metricstest.L{"type": "tap", "seat_kind": "agent", "result": "applied"}},
		{"ApplyAction admin bound to a seat", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: actions.TypeChangeLife, Actor: human, Admin: true}, ok)
			return err
		}, metricstest.L{"type": "change_life", "seat_kind": "admin", "result": "applied"}},
		{"ApplyAction admin with no seat", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: actions.TypeMoveCard}, ok)
			return err
		}, metricstest.L{"type": "move_card", "seat_kind": "admin", "result": "applied"}},
		{"ApplyAction unknown type", func() error {
			_, _, err := room.ApplyAction(ActionTag{Type: "bundle", Actor: human}, refuse)
			return err
		}, metricstest.L{"type": "other", "seat_kind": "human", "result": "rejected"}},
		{"ApplyExternalAction", func() error {
			_, _, err := room.ApplyExternalAction(ActionTag{Type: actions.TypeRollTableDie, Actor: bot}, ok)
			return err
		}, metricstest.L{"type": "roll_table_die", "seat_kind": "bot", "result": "applied"}},
		{"ApplyExternalAction refused", func() error {
			_, _, err := room.ApplyExternalAction(ActionTag{Type: actions.TypeSetTableSettings, Actor: human}, refuse)
			return err
		}, metricstest.L{"type": "set_table_settings", "seat_kind": "human", "result": "rejected"}},
		{"ApplyExternal", func() error { _, _, err := room.ApplyExternal(ok); return err },
			metricstest.L{"type": "lobby", "seat_kind": "human", "result": "applied"}},
		{"ApplyExternal refused", func() error { _, _, err := room.ApplyExternal(refuse); return err },
			metricstest.L{"type": "lobby", "seat_kind": "human", "result": "rejected"}},
		{"ApplyBundle", func() error {
			_, _, err := room.ApplyBundle(Bundle{Actor: bot, Steps: []func() error{ok, ok, ok}})
			return err
		}, metricstest.L{"type": "bundle", "seat_kind": "bot", "result": "applied"}},
		{"ApplyBundle refused", func() error {
			_, _, err := room.ApplyBundle(Bundle{Actor: human, Steps: []func() error{ok, refuse}})
			return err
		}, metricstest.L{"type": "bundle", "seat_kind": "human", "result": "rejected"}},
		{"ApplyBundle admin", func() error {
			_, _, err := room.ApplyBundle(Bundle{Steps: []func() error{ok}})
			return err
		}, metricstest.L{"type": "bundle", "seat_kind": "admin", "result": "applied"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			all := actionsTotal(t, nil)
			series := actionsTotal(t, c.want)
			held := applySeconds(t, c.want["result"])
			err := c.commit()
			if (err != nil) != (c.want["result"] == "rejected") {
				t.Fatalf("commit err = %v, want result %s", err, c.want["result"])
			}
			if got := actionsTotal(t, c.want) - series; got != 1 {
				t.Errorf("cmdctrl_actions_total%v rose by %v, want 1", c.want, got)
			}
			if got := actionsTotal(t, nil) - all; got != 1 {
				t.Errorf("cmdctrl_actions_total rose by %v in all, want 1", got)
			}
			if got := applySeconds(t, c.want["result"]) - held; got != 1 {
				t.Errorf("cmdctrl_action_apply_seconds{result=%s} observed %v, want 1", c.want["result"], got)
			}
		})
	}
}

// Undo is not one of the three paths, and counts nothing.
func TestUndoIsNotAnAction(t *testing.T) {
	room, human, _, _ := seatKindsRoom(t)
	if _, _, err := room.ApplyAction(ActionTag{Type: actions.TypeTap, Actor: human}, ok); err != nil {
		t.Fatal(err)
	}
	before := actionsTotal(t, nil)
	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if got := actionsTotal(t, nil) - before; got != 0 {
		t.Errorf("an undo moved cmdctrl_actions_total by %v", got)
	}
}

// Every type in the actions enum is its own label value; nothing in it
// falls to "other".
func TestEveryActionTypeIsALabel(t *testing.T) {
	registered := map[string]bool{}
	for _, v := range metrics.ActionTypes() {
		registered[v] = true
	}
	for _, ty := range actions.Types() {
		if got := (ActionTag{Type: ty}).typeLabel(); got != string(ty) {
			t.Errorf("type %q is labelled %q", ty, got)
		}
		if !registered[string(ty)] {
			t.Errorf("type %q is not registered with the metrics package", ty)
		}
	}
}

// --- restore points ---------------------------------------------------

// persistedRoom is an active, persisted room with the given live seq
// and restore-point bookkeeping.
func persistedRoom(t *testing.T, seq, restoreSeq uint64, at time.Time) *Room {
	t.Helper()
	r := NewRoom(newPersistGame(t), quietLog(), t.TempDir())
	r.seq = seq
	r.lastRestorePoint = restorePointRecord{Seq: restoreSeq, At: at}
	return r
}

func TestRestorePointLag(t *testing.T) {
	now := time.Now()
	mgr := NewRoomManager(quietLog(), t.TempDir())

	// At its restore point, however old: a deploy rewinds nothing.
	mgr.Register(persistedRoom(t, 7, 7, now.Add(-time.Hour)))
	// Two commits past a 90 s old restore point.
	mgr.Register(persistedRoom(t, 9, 7, now.Add(-90*time.Second)))
	// One commit past a 10 s old one.
	mgr.Register(persistedRoom(t, 4, 3, now.Add(-10*time.Second)))

	// Behind, but a practice table: it writes no restore points.
	practice := persistedRoom(t, 9, 1, now.Add(-2*time.Hour))
	practice.dumpDir = ""
	mgr.Register(practice)

	// Behind, but not active: a lobby-state table.
	lobby := NewRoom(game.NewGame(), quietLog(), t.TempDir())
	lobby.seq = 3
	mgr.Register(lobby)

	got := mgr.RestorePointLag(now)
	if got.Behind != 2 {
		t.Errorf("Behind = %d, want 2", got.Behind)
	}
	if got.Oldest != 90*time.Second {
		t.Errorf("Oldest = %v, want 90s", got.Oldest)
	}

	// A room that has never written one is aged from its game's
	// creation: a deploy would lose all of it.
	never := persistedRoom(t, 5, 0, time.Time{})
	never.Game.CreatedAt = now.Add(-3 * time.Hour)
	mgr.Register(never)
	got = mgr.RestorePointLag(now)
	if got.Behind != 3 || got.Oldest != 3*time.Hour {
		t.Errorf("with a room that never wrote one: %+v, want 3 behind, oldest 3h", got)
	}
}

// A real room: a clean commit writes its restore point, so the room is
// not behind afterwards.
func TestRestorePointLagOnARealCommit(t *testing.T) {
	mgr := NewRoomManager(quietLog(), t.TempDir())
	g := newPersistGame(t)
	room := mgr.Create(g)
	seat := g.Seats[0].ID
	if _, _, err := room.ApplyAction(ActionTag{Type: actions.TypeChangeLife, Actor: seat}, func() error {
		g.WithWriteLock(func() { g.Seats[0].Life = 33 })
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := mgr.RestorePointLag(time.Now()); got.Behind != 0 || got.Oldest != 0 {
		t.Errorf("after a clean commit: %+v, want nothing behind", got)
	}
}

// Scraping the restore-point collector while seats commit is race-free
// and deadlock-free: the collector never holds two locks (run under
// -race).
func TestRestorePointCollectorScrapesWhileApplying(t *testing.T) {
	mgr := NewRoomManager(quietLog(), t.TempDir())
	var rooms []*Room
	for i := 0; i < 3; i++ {
		rooms = append(rooms, mgr.Create(newPersistGame(t)))
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(metrics.NewRestorePointCollector(mgr))

	var wg sync.WaitGroup
	for _, r := range rooms {
		wg.Add(1)
		go func(r *Room) {
			defer wg.Done()
			g := r.Game
			seat := g.Seats[0].ID
			for i := 0; i < 40; i++ {
				_, _, _ = r.ApplyAction(ActionTag{Type: actions.TypeChangeLife, Actor: seat}, func() error {
					g.WithWriteLock(func() { g.Seats[0].Life-- })
					return nil
				})
			}
		}(r)
	}
	stop := make(chan struct{})
	scraped := make(chan int)
	go func() {
		n := 0
		for {
			select {
			case <-stop:
				scraped <- n
				return
			default:
			}
			if _, err := reg.Gather(); err != nil {
				t.Errorf("gather: %v", err)
			}
			n++
		}
	}()
	wg.Wait()
	close(stop)
	if n := <-scraped; n == 0 {
		t.Error("the collector was never scraped")
	}
}

// --- boot ---------------------------------------------------------------

func TestRestoreSummarySetsTheBootGauges(t *testing.T) {
	room := NewRoom(game.NewGame(), quietLog(), "")
	outcomes := []RestoreOutcome{
		{Room: room, AbilityShortfalls: []game.AbilityShortfall{{}, {}}},
		{Room: room},
		{Skipped: "game already ended"},
		{Err: errors.New("bad"), Reason: ReasonDecodeError},
		{Err: errors.New("new"), Reason: ReasonSchemaTooNew, LostStackAbilities: []game.LostStackAbility{{}}},
	}
	LogRestoreSummary(quietLog(), outcomes)
	gauge := func(outcome string) float64 {
		return metricstest.Value(t, "cmdctrl_boot_restore_games", metricstest.L{"outcome": outcome})
	}
	if gauge("restored") != 2 || gauge("ended") != 1 || gauge("abandoned") != 2 {
		t.Errorf("boot restore games = restored %v, ended %v, abandoned %v; want 2, 1, 2",
			gauge("restored"), gauge("ended"), gauge("abandoned"))
	}
	if got := metricstest.Value(t, "cmdctrl_boot_restore_degraded_cards", nil); got != 3 {
		t.Errorf("degraded cards = %v, want 3", got)
	}

	// A boot with nothing to restore sets them to zero.
	LogRestoreSummary(quietLog(), nil)
	for _, o := range []string{"restored", "ended", "abandoned"} {
		if g := gauge(o); g != 0 {
			t.Errorf("%s = %v after an empty restore pass, want 0", o, g)
		}
	}
	if err := metrics.CheckClosedLabels(metrics.Registry); err != nil {
		t.Error(err)
	}
}
