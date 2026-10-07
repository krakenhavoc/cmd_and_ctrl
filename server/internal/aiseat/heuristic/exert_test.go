package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exertAttackMove is attackMove's twin that also exerts the attacker
// (ADR 0130 §6).
func exertAttackMove(t *testing.T, seat int, attacker string, target int) legal.Move {
	m := attackMove(t, seat, attacker, target)
	m.Label += " and exert it (it won't untap during your next untap step)"
	m.Params = mustJSON(t, map[string]any{"attacker": attacker, "target": seatID(target).String(), "exert": true})
	return m
}

// TestHeuristicNeverPicksAnExertMove is ADR 0130 test 11's bot half: in
// PR 1 the heuristic attacks exactly as before and ignores the twin
// move that exerts (pricing an exert is PR 3, §9). The exert move is
// listed FIRST, so a policy that read the list blindly would take it.
func TestHeuristicNeverPicksAnExertMove(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(12)), newSeat(1, withLife(2))},
		withBattlefield(creature(cardID(10), 0, "Avenger", 3, 1)),
		withTurn(9, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0), exertAttackMove(t, 0, cardID(10), 1), attackMove(t, 0, cardID(10), 1))
	got := chose(t, in, decide(t, heuristic.New(), in))
	if got != attackMove(t, 0, cardID(10), 1).Label {
		t.Fatalf("chose %q, want the plain attack", got)
	}

	// With only the exert move on offer the heuristic doesn't attack.
	in = input(0, v, passMove(0), exertAttackMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q; the heuristic never exerts in PR 1", got)
	}
}
