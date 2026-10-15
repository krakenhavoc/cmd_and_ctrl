package boardtext_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
)

// turns_test.go — ADR 0188 Decision 10 (#3566): the golden lines a model
// or the MCP seat reads for an added combat, a turn skip and whose turn
// is next.

func TestRenderNamesAnAddedPhaseASkipAndTheNextTurn(t *testing.T) {
	v := view()
	v.Turn.Step = "declare_attackers"
	v.Turn.Phase = "combat"
	v.Turn.PhaseOrdinal = 2
	next := 1
	v.Turn.NextSeat = &next
	v.Seats[1].TurnSkips = 1
	got := boardtext.Render(v, "a", boardtext.Options{})
	for _, want := range []string{
		"TURN 5 (round 3) — declare attackers (combat 2) — active player: Ann (YOU) — priority: Ann (YOU) — next turn: Bo\n",
		"Bo — 38 life, 0 cards in hand, 0 in library, skips their next turn\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}

	// Two skips, a third main phase, an extra turn and the bot's own seat
	// next.
	v = view()
	v.Turn.Step = "postcombat_main"
	v.Turn.Phase = "postcombat_main"
	v.Turn.PhaseOrdinal = 3
	v.Turn.Extra = true
	zero := 0
	v.Turn.NextSeat = &zero
	v.Seats[0].TurnSkips = 2
	got = boardtext.Render(v, "a", boardtext.Options{})
	for _, want := range []string{
		"TURN 5 (round 3) — postcombat main (main 3) (extra turn) — active player: Ann (YOU) — priority: Ann (YOU) — next turn: Ann (YOU)\n",
		"Ann (YOU) — 40 life, 1 cards in hand, 0 in library, skips their next 2 turns\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
}

// TestRenderReadsThePhaseFamilyFromTheStepWhenPhaseIsAbsent: a view that
// carries phase_ordinal but not phase still names the family.
func TestRenderReadsThePhaseFamilyFromTheStepWhenPhaseIsAbsent(t *testing.T) {
	v := view()
	v.Turn.Step = "end_combat"
	v.Turn.PhaseOrdinal = 2
	got := boardtext.Render(v, "a", boardtext.Options{})
	if !strings.Contains(got, "end combat (combat 2)") {
		t.Errorf("no combat ordinal in\n%s", got)
	}
}

// TestRenderLeavesOrdinaryTurnsAlone: the first combat of a turn, no skip
// and no next_seat add nothing to the lines.
func TestRenderLeavesOrdinaryTurnsAlone(t *testing.T) {
	v := view()
	v.Turn.Phase = "precombat_main"
	v.Turn.PhaseOrdinal = 1
	got := boardtext.Render(v, "a", boardtext.Options{})
	for _, bad := range []string{"(main", "(combat", "next turn:", "skips"} {
		if strings.Contains(got, bad) {
			t.Errorf("an ordinary turn mentions %q in\n%s", bad, got)
		}
	}
}
