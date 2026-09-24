package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cant_lose_test.go — ADR 0057 Decision 6 (#749), test plan item 18:
// the heuristic reads each clock against the seat's cant_lose. A seat
// that can't lose to its life total is not hopeless, is not a lethal
// target, and is not in more danger at -10 than at 0.

func withCantLose(causes ...string) seatOpt {
	return func(p *protocol.PlayerView) { p.CantLose = causes }
}

// TestNoConcedeBehindACantLoseGate: the concede fixture, three turns
// running, with the bot behind its own Platinum Angel. Conceding is
// the one loss its gate can't stop, so it must never be the one to
// take it.
func TestNoConcedeBehindACantLoseGate(t *testing.T) {
	p := heuristic.New()
	for turn := 1; turn <= 6; turn++ {
		if p.ShouldConcede(hopelessAt(turn, withLife(-10), withCantLose("life", "empty_draw", "poison", "commander_damage", "effect"))) {
			t.Fatalf("conceded on turn %d at -10 life behind a can't-lose gate", turn)
		}
	}
	// The control: the same position without the gate concedes.
	q := heuristic.New()
	conceded := false
	for turn := 1; turn <= 6 && !conceded; turn++ {
		conceded = q.ShouldConcede(hopelessAt(turn, withLife(-10)))
	}
	if !conceded {
		t.Fatalf("the control never conceded; the fixture is not hopeless")
	}
}

// TestLethalPushSkipsASeatThatCantLoseToLife: the lethal board from
// TestLethalPushCountsOnlyTheBlocksThatExist, but the defender can't
// lose to its life total. No swing is lethal there.
func TestLethalPushSkipsASeatThatCantLoseToLife(t *testing.T) {
	flier := keywords("flying")
	board := func(opts ...seatOpt) (d string) {
		v := newView(
			[]protocol.PlayerView{newSeat(0, withLife(6)), newSeat(1, append([]seatOpt{withLife(3)}, opts...)...)},
			withBattlefield(
				creature(cardID(10), 0, "Drake", 3, 3, flier),
				creature(cardID(11), 0, "Drake", 3, 3, flier),
				creature(cardID(12), 0, "Drake", 3, 3, flier),
				creature(cardID(20), 1, "Ogre", 4, 4),
			),
			withTurn(20, 0, "declare_attackers"),
		)
		in := input(0, v, passMove(0),
			attackMove(t, 0, cardID(10), 1), attackMove(t, 0, cardID(11), 1), attackMove(t, 0, cardID(12), 1),
		)
		return decide(t, heuristic.New(), in).Reason
	}
	if r := board(); !strings.Contains(r, "all-in for the kill") {
		t.Fatalf("control: reason %q, want the all-in push", r)
	}
	if r := board(withCantLose("life")); strings.Contains(r, "all-in for the kill") {
		t.Fatalf("pushed all-in (%q) at a seat that can't lose to damage", r)
	}
}

// TestLifeDangerIsClampedBehindAGate: a seat at -10 behind a gate is
// scored like a seat at 0, not like one twenty points deeper into the
// squared penalty.
func TestLifeDangerIsClampedBehindAGate(t *testing.T) {
	w := heuristic.DefaultWeights()
	at := func(life int, opts ...seatOpt) float64 {
		v := newView([]protocol.PlayerView{newSeat(0, append([]seatOpt{withLife(life)}, opts...)...), newSeat(1)})
		return w.Evaluate(v)[seatID(0).String()].Strength
	}
	gated := withCantLose("life")
	zero, deep := at(0, gated), at(-10, gated)
	// Only the linear life term separates them now.
	if want := w.Life * 10; zero-deep > want+1e-9 {
		t.Errorf("gated -10 scores %.2f below gated 0, want at most the linear %.2f", zero-deep, want)
	}
	if at(0)-at(-10) <= zero-deep {
		t.Errorf("the ungated seat is not penalised more for -10 than the gated one")
	}
}
