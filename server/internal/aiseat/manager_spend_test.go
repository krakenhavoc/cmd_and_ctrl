package aiseat_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
)

// manager_spend_test.go is #735's wiring half: the game's ONE spend
// record — built when every seat is done, logged as the admin summary
// line and written into the decision log before it closes.
//
// The unit half (what a funnel counts, and the decision/improvisation
// split) is aiseat/model/spend_test.go; the whole-game half (four
// model seats against a fake client, totals end to end) is
// spend_game_test.go. This file asks the question neither of those
// can: does the number ever leave the process, once, per game.

// --- doubles ---------------------------------------------------------

// spendPolicy is a seat that costs a known amount and plays nothing.
// It declines every window, which parks the runner, so the table is
// quiescent and the assertions are about the manager rather than
// about a game.
type spendPolicy struct {
	name  string
	spend aiseat.Spend
}

func (p *spendPolicy) Name() string { return p.name }

func (p *spendPolicy) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	return aiseat.Decision{Index: aiseat.Decline, Reason: "test"}, nil
}

func (p *spendPolicy) Spend() aiseat.Spend { return p.spend }

// spendFactory seats spendPolicy at every tier.
type spendFactory struct {
	spend aiseat.Spend
}

func (f spendFactory) NewPolicy(seat aiseat.SeatSpec) (aiseat.Policy, error) {
	return &spendPolicy{name: seat.Tier, spend: f.spend}, nil
}

func (spendFactory) TierStatus(aiseat.Tier) aiseat.TierStatus {
	return aiseat.TierStatus{Available: true}
}

func (spendFactory) RunnerConfig(aiseat.Tier) aiseat.Config { return aiseat.Config{} }

// spendLog is a decision log that records the spend record and the
// order it arrived in relative to Close. The order is the point: a
// record written after the file is closed is a record nobody has.
type spendLog struct {
	mu            sync.Mutex
	spends        []aiseat.GameSpend
	closes        int
	closedBefore  bool
	observedAfter bool
}

func (l *spendLog) Observe(aiseat.DecisionEvent) {}

func (l *spendLog) ObserveSpend(gs aiseat.GameSpend) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closes > 0 {
		l.observedAfter = true
	}
	l.spends = append(l.spends, gs)
}

func (l *spendLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.spends) == 0 {
		l.closedBefore = true
	}
	l.closes++
	return nil
}

func (l *spendLog) snapshot() ([]aiseat.GameSpend, int, bool, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]aiseat.GameSpend(nil), l.spends...), l.closes, l.closedBefore, l.observedAfter
}

type spendLogger struct{ log *spendLog }

func (d spendLogger) OpenGame(uuid.UUID) (aiseat.GameDecisionLog, error) { return d.log, nil }

// capturingHandler keeps every log line's message and attributes, so
// a test can assert on the admin summary rather than on stdout.
type capturingHandler struct {
	mu    sync.Mutex
	lines []loggedLine
}

type loggedLine struct {
	msg   string
	attrs map[string]slog.Value
}

func (h *capturingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *capturingHandler) Handle(_ context.Context, r slog.Record) error {
	line := loggedLine{msg: r.Message, attrs: map[string]slog.Value{}}
	r.Attrs(func(a slog.Attr) bool {
		line.attrs[a.Key] = a.Value
		return true
	})
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lines = append(h.lines, line)
	return nil
}

func (h *capturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *capturingHandler) WithGroup(string) slog.Handler { return h }

func (h *capturingHandler) withMessage(msg string) []loggedLine {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []loggedLine
	for _, l := range h.lines {
		if l.msg == msg {
			out = append(out, l)
		}
	}
	return out
}

// --- the tests -------------------------------------------------------

const adminSpendLine = "bot model spend for the game"

func seatSpend(decisionCalls, improvCalls int64, in, out int) aiseat.Spend {
	return aiseat.Spend{
		Decision: aiseat.PurposeSpend{
			Calls: decisionCalls,
			Usage: aiseat.TokenUsage{InputTokens: in, OutputTokens: out},
		},
		Improvisation: aiseat.PurposeSpend{
			Calls: improvCalls,
			Usage: aiseat.TokenUsage{InputTokens: 10 * in, OutputTokens: 10 * out},
		},
	}
}

// One record per game: every seat in it, the totals added, logged as
// the admin line and handed to the decision log before it closes.
func TestAGameProducesOneSpendRecord(t *testing.T) {
	room := newRoom(t, 2, 7351)
	per := seatSpend(11, 2, 100, 5)
	dlog := &spendLog{}
	logs := &capturingHandler{}

	mgr := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, slog.New(logs))
	mgr.SetPolicyFactory(spendFactory{spend: per})
	mgr.SetDecisionLogger(spendLogger{log: dlog})
	mgr.StartBots(room, []aiseat.SeatSpec{
		{PlayerID: room.Game.Seats[0].ID, Tier: string(aiseat.TierAssisted)},
		{PlayerID: room.Game.Seats[1].ID, Tier: string(aiseat.TierAssisted)},
	})
	mgr.StopBots(room.Game.ID)

	spends, closes, closedBefore, observedAfter := dlog.snapshot()
	if len(spends) != 1 {
		t.Fatalf("the decision log got %d spend records for one game; it is one per table", len(spends))
	}
	if closes != 1 {
		t.Errorf("the decision log was closed %d times", closes)
	}
	if closedBefore || observedAfter {
		t.Error("the spend record was written after the log was closed — nobody would ever read it")
	}

	gs := spends[0]
	if gs.Game != room.Game.ID {
		t.Errorf("spend record names game %s, want %s", gs.Game, room.Game.ID)
	}
	if len(gs.Seats) != 2 {
		t.Fatalf("spend record has %d seats, want 2", len(gs.Seats))
	}
	for _, s := range gs.Seats {
		if s.Policy != string(aiseat.TierAssisted) {
			t.Errorf("seat %s reported policy %q", s.Seat, s.Policy)
		}
		if s.Spend != per {
			t.Errorf("seat %s reported %+v, want %+v", s.Seat, s.Spend, per)
		}
	}
	total := gs.Total()
	if total.Decision.Calls != 22 || total.Improvisation.Calls != 4 {
		t.Errorf("table total = %d decision / %d improvisation calls, want 22 / 4",
			total.Decision.Calls, total.Improvisation.Calls)
	}
	if got, want := total.Total().Usage.InputTokens, 2*(100+1000); got != want {
		t.Errorf("table total input tokens = %d, want %d", got, want)
	}

	// The admin summary: one line, with the numbers on it.
	lines := logs.withMessage(adminSpendLine)
	if len(lines) != 1 {
		t.Fatalf("the game produced %d admin spend lines, want exactly 1", len(lines))
	}
	for key, want := range map[string]int64{
		"calls": 26, "decision_calls": 22, "improv_calls": 4,
		"input_tokens": 2200, "output_tokens": 110, "seats": 2,
	} {
		v, ok := lines[0].attrs[key]
		if !ok {
			t.Errorf("the admin spend line has no %q attribute", key)
			continue
		}
		if v.Int64() != want {
			t.Errorf("admin spend line %s = %d, want %d", key, v.Int64(), want)
		}
	}
	if got := lines[0].attrs["tiers"].String(); got != "assisted x2" {
		t.Errorf("admin spend line tiers = %q, want %q", got, "assisted x2")
	}

	// Idempotent: a game is stopped once, and the second stop is a
	// no-op rather than a second bill.
	mgr.StopBots(room.Game.ID)
	if spends, _, _, _ := dlog.snapshot(); len(spends) != 1 {
		t.Errorf("a second StopBots produced another spend record: %d", len(spends))
	}
	if n := len(logs.withMessage(adminSpendLine)); n != 1 {
		t.Errorf("a second StopBots logged another admin spend line: %d", n)
	}
}

// A game that simply ENDS is never stopped — the runners see the
// state leave StateActive and return on their own — so the record has
// to be produced by the watcher, not by StopBots.
//
// Before #735 the watcher only ran for a game with a decision log,
// because closing that file was the only thing it had to do. A spend
// record that only a DELETED game produced would miss almost every
// game played.
func TestAFinishedGameStillProducesItsSpendRecord(t *testing.T) {
	room := newRoom(t, 2, 7352)
	logs := &capturingHandler{}
	mgr := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, slog.New(logs))
	mgr.SetPolicyFactory(spendFactory{spend: seatSpend(3, 0, 50, 2)})
	// No decision logger at all: the admin line is the only surface,
	// and it is the one an operator actually has.
	mgr.StartBots(room, []aiseat.SeatSpec{
		{PlayerID: room.Game.Seats[0].ID, Tier: string(aiseat.TierHeuristic)},
		{PlayerID: room.Game.Seats[1].ID, Tier: string(aiseat.TierHeuristic)},
	})
	runners := mgr.Runners(room.Game.ID)
	if len(runners) != 2 {
		t.Fatalf("runners: %d", len(runners))
	}
	// End the game without StopBots, and commit so the runners wake
	// and see it — which is exactly how a finished table clears.
	room.Game.End()
	if _, _, err := room.ApplyExternal(func() error { return nil }); err != nil {
		t.Fatalf("commit the ending: %v", err)
	}
	for _, r := range runners {
		select {
		case <-r.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("a runner did not exit when the game ended")
		}
	}
	waitFor(t, "the game's spend record", func() bool {
		return len(logs.withMessage(adminSpendLine)) == 1
	})
	line := logs.withMessage(adminSpendLine)[0]
	if got := line.attrs["decision_calls"].Int64(); got != 6 {
		t.Errorf("decision_calls = %d, want 6", got)
	}
	mgr.Shutdown()
	if n := len(logs.withMessage(adminSpendLine)); n != 1 {
		t.Errorf("shutdown logged the finished game's spend a second time: %d lines", n)
	}
}

// A table with no model on it still gets its line. "This game cost
// nothing" is a measurement, and a line that appears only when there
// is a bill cannot be told from a line that failed to be written.
func TestATableWithNoModelStillReportsZero(t *testing.T) {
	room := newRoom(t, 2, 7353)
	logs := &capturingHandler{}
	mgr := aiseat.NewManagerWithConfig(nil, aiseat.Config{MinThink: time.Hour}, slog.New(logs))
	mgr.StartBots(room, []aiseat.SeatSpec{
		{PlayerID: room.Game.Seats[0].ID, Tier: string(aiseat.TierRandom)},
		{PlayerID: room.Game.Seats[1].ID, Tier: string(aiseat.TierRandom)},
	})
	mgr.StopBots(room.Game.ID)
	lines := logs.withMessage(adminSpendLine)
	if len(lines) != 1 {
		t.Fatalf("a random table produced %d spend lines, want 1", len(lines))
	}
	if got := lines[0].attrs["calls"].Int64(); got != 0 {
		t.Errorf("a random table reported %d model calls", got)
	}
}
