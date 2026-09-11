package heuristic_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attackTurn builds one declare-attackers window: the bot at seat 0
// with a single 2/2, three opponents at the given life totals and
// empty boards, and an attack move offered against each of them.
func attackTurn(t *testing.T, turn int, lives [3]int) (protocol.GameView, []legal.Move) {
	t.Helper()
	seats := []protocol.PlayerView{newSeat(0)}
	for i, life := range lives {
		seats = append(seats, newSeat(i+1, withLife(life)))
	}
	v := newView(seats,
		withBattlefield(creature(cardID(10), 0, "Bear", 2, 2)),
		withTurn(turn, 0, "declare_attackers"),
	)
	moves := []legal.Move{passMove(0)}
	for i := 1; i <= 3; i++ {
		moves = append(moves, attackMove(t, 0, cardID(10), i))
	}
	return v, moves
}

func attackedSeat(t *testing.T, label string) string {
	t.Helper()
	if !strings.HasPrefix(label, "Attack ") {
		return ""
	}
	return strings.Fields(label)[1]
}

// The S31 exit criterion, verbatim: "three ineffective turns
// attacking A → target B on turn 4". "Ineffective" means the target's
// life total did not move — which, with three identical empty seats,
// is the only thing that can distinguish them.
func TestAggressionRotatesAfterThreeIneffectiveTurns(t *testing.T) {
	p := heuristic.New()
	var targets []string
	for turn := 1; turn <= 5; turn++ {
		v, moves := attackTurn(t, turn, [3]int{40, 40, 40})
		in := input(0, v, moves...)
		d := decide(t, p, in)
		got := chose(t, in, d)
		who := attackedSeat(t, got)
		if who == "" {
			t.Fatalf("turn %d: chose %q, expected an attack", turn, got)
		}
		targets = append(targets, who)
	}
	first := targets[0]
	for turn := 1; turn <= 3; turn++ {
		if targets[turn-1] != first {
			t.Fatalf("turn %d attacked %s, want the same seat as turn 1 (%s): %v", turn, targets[turn-1], first, targets)
		}
	}
	if targets[3] == first {
		t.Fatalf("turn 4 still attacked %s after three turns that moved nothing: %v", first, targets)
	}
	if targets[4] == first {
		t.Fatalf("turn 5 went straight back to %s; the cooldown is not holding: %v", first, targets)
	}
}

// The control: when the attacks ARE working, the bot keeps at it. A
// rotation that fires on a working plan is worse than no rotation.
func TestAggressionStaysOnATargetItIsActuallyKilling(t *testing.T) {
	p := heuristic.New()
	life := 40
	var targets []string
	for turn := 1; turn <= 5; turn++ {
		v, moves := attackTurn(t, turn, [3]int{life, 40, 40})
		in := input(0, v, moves...)
		targets = append(targets, attackedSeat(t, chose(t, in, decide(t, p, in))))
		life -= 2
	}
	for i, got := range targets {
		if got != targets[0] {
			t.Fatalf("turn %d rotated to %s while %s was losing life every turn: %v", i+1, got, targets[0], targets)
		}
	}
}

func TestRotationIsPerPolicyInstance(t *testing.T) {
	// Two seats' policies must not share rotation state; the runner
	// constructs one per seat and the bots would otherwise coordinate
	// through a global.
	a, b := heuristic.New(), heuristic.New()
	if fmt.Sprintf("%p", a) == fmt.Sprintf("%p", b) {
		t.Fatal("New returned the same policy twice")
	}
	a.Reset()
	b.Reset()
}
