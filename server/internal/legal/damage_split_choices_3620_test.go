package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// damage_split_choices_3620_test.go — #3620. The enumerator used to
// offer one damage split, so a text seat (MCP or bot) could never pick
// which blockers die; the aiseat/rules classifier called the window
// forced and the MCP seat absorbed it. CR 510.1c-d let the attacker
// divide as it likes, so the enumerator now offers the canonical split
// first and then each maximal set of blockers that can all take lethal.

// splitsOf decodes every offered move into its per-blocker amounts.
func splitsOf(t *testing.T, moves []legal.Move) []map[uuid.UUID]int {
	t.Helper()
	out := make([]map[uuid.UUID]int, len(moves))
	for i, m := range moves {
		out[i], _ = splitOf(t, m)
	}
	return out
}

// The report: a 6/6 blocked by two 4/4s and a 3/3. The canonical split
// kills a 4/4 (2 left over on the 3/3); the player wants the 3/3 dead
// (the commander) and the other 3 on a 4/4.
func TestDamageSplitsOfferEachMaximalKillSet(t *testing.T) {
	g, me, ids := blockedCombat(t, 6, 6, nil, [][2]int{{4, 4}, {4, 4}, {3, 3}})
	gatta, luzzu, tidus := ids[0], ids[1], ids[2]
	moves := assignmentMoves(t, g, me)
	splits := splitsOf(t, moves)
	if len(moves) < 2 {
		t.Fatalf("%d moves offered, want the canonical split and the others", len(moves))
	}
	// Canonical first: it kills a 4/4 (worth more than the 3/3) and puts
	// the 2 left over on the other 4/4.
	if c := splits[0]; c[gatta] != 4 || c[luzzu] != 2 || c[tidus] != 0 {
		t.Errorf("canonical split = %v, want 4 to the first 4/4 and 2 to the second", c)
	}
	var tidusDies bool
	for i, sp := range splits {
		if sp[tidus] == 3 && sp[gatta]+sp[luzzu] == 3 {
			tidusDies = true
			if !strings.Contains(moves[i].Label, "3 to Blocker (lethal)") {
				t.Errorf("label %q does not say the 3/3 takes lethal", moves[i].Label)
			}
		}
		total := 0
		for _, n := range sp {
			total += n
		}
		if total != 6 {
			t.Errorf("split %v totals %d, want 6", sp, total)
		}
	}
	if !tidusDies {
		t.Fatalf("no offered split kills the 3/3 and puts the other 3 elsewhere: %v", splits)
	}
	// The two 4/4s are the same creature to the attacker: killing the
	// 3/3 offers one survivor to take the remainder, not both.
	n := 0
	for _, sp := range splits {
		if sp[tidus] == 3 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d splits kill the 3/3, want 1 (identical survivors count once)", n)
	}
	// Every offer is something the engine accepts.
	dispatchAll(t, g, me, moves)
}

// Answering the second offer in the report kills the commander.
func TestDamageSplitChosenKillsTheOtherBlocker(t *testing.T) {
	g, me, ids := blockedCombat(t, 6, 6, nil, [][2]int{{4, 4}, {4, 4}, {3, 3}})
	tidus := ids[2]
	var pick *legal.Move
	moves := assignmentMoves(t, g, me)
	for i := range moves {
		if sp, _ := splitOf(t, moves[i]); sp[tidus] == 3 {
			pick = &moves[i]
		}
	}
	if pick == nil {
		t.Fatal("no split gives the 3/3 lethal")
	}
	dispatchAll(t, g, me, []legal.Move{*pick})
	clone := g.Clone()
	if err := actions.Dispatch(clone, actions.Action{
		Type: actions.Type(pick.Type), Player: pick.Player, Caller: me, Params: pick.Params,
	}); err != nil {
		t.Fatal(err)
	}
	if c := onBattlefield(clone, tidus); c != nil {
		t.Errorf("the 3/3 survived with %d damage", c.DamageMarked)
	}
}

// Power enough for every blocker is no choice: one split, and a
// trampler's remainder goes over (#2956).
func TestDamageSplitsOfferOneWhenPowerCoversLethal(t *testing.T) {
	g, me, _ := blockedCombat(t, 9, 9, []string{"trample"}, [][2]int{{4, 4}, {3, 3}})
	if moves := assignmentMoves(t, g, me); len(moves) != 1 {
		t.Fatalf("%d moves offered when the power covers every blocker, want 1", len(moves))
	}
}

// A trampler short of lethal everywhere sends nothing over, in every
// offered split (CR 702.19b).
func TestDamageSplitsNeverTrampleShortOfLethal(t *testing.T) {
	g, me, _ := blockedCombat(t, 6, 6, []string{"trample"}, [][2]int{{4, 4}, {4, 4}, {3, 3}})
	moves := assignmentMoves(t, g, me)
	if len(moves) < 2 {
		t.Fatalf("%d moves offered, want several", len(moves))
	}
	for _, m := range moves {
		if _, tr := splitOf(t, m); tr != 0 {
			t.Errorf("%q tramples %d over blockers it has not killed", m.Label, tr)
		}
	}
	dispatchAll(t, g, me, moves)
}

// The offer is bounded: a wide block keeps the canonical split alone,
// and a narrower one stays under maxDamageSplits.
func TestDamageSplitsAreBounded(t *testing.T) {
	// Nine blockers is past maxDamageSplitBlockers.
	blockers := make([][2]int, 9)
	for i := range blockers {
		blockers[i] = [2]int{1, 3 + i}
	}
	g, me, _ := blockedCombat(t, 5, 5, nil, blockers)
	if moves := assignmentMoves(t, g, me); len(moves) != 1 {
		t.Errorf("%d moves for nine blockers, want the canonical one", len(moves))
	}
	// Eight distinct blockers that all fit pairwise: many maximal sets.
	blockers = blockers[:8]
	g, me, _ = blockedCombat(t, 13, 13, nil, blockers)
	moves := assignmentMoves(t, g, me)
	if len(moves) < 2 || len(moves) > 12 {
		t.Errorf("%d moves for eight blockers, want 2..12", len(moves))
	}
	dispatchAll(t, g, me, moves)
}

// The canonical split stays first and is the one CanonicalDamageSplit
// returns, which the view and the bots depend on.
func TestDamageSplitsLeadWithTheCanonicalSplit(t *testing.T) {
	g, me, _ := blockedCombat(t, 6, 6, nil, [][2]int{{4, 4}, {4, 4}, {3, 3}})
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceDamageAssignment {
			c = pc
		}
	}
	if c == nil {
		t.Fatal("no damage-assignment prompt")
	}
	want, ok := legal.CanonicalDamageSplit(g, c)
	if !ok {
		t.Fatal("no canonical split")
	}
	got, _ := splitOf(t, assignmentMoves(t, g, me)[0])
	for _, a := range want.Assignments {
		if got[a.BlockerID] != a.Amount {
			t.Errorf("first move = %v, want the canonical split %v", got, want.Assignments)
			break
		}
	}
}

// The window is no longer forced, so Layer A (and the MCP seat's
// default absorb rules, which absorb "forced") leaves it to the seat.
// A one-split window is still forced: nothing is lost by taking it.
func TestDamageWindowWithSeveralSplitsIsNotForced(t *testing.T) {
	resolve := func(g *game.Game, seat uuid.UUID) rules.Verdict {
		return rules.Resolve(aiseat.Input{
			View:  protocol.ViewOfGameFor(g, seat.String()),
			Seat:  seat,
			Moves: legal.EnumerateFor(g, seat),
		})
	}
	g, me, _ := blockedCombat(t, 6, 6, nil, [][2]int{{4, 4}, {4, 4}, {3, 3}})
	if v := resolve(g, me); v.Absorbed() {
		t.Errorf("a window with several splits was absorbed as %q", v.Rule)
	}
	g, me, _ = blockedCombat(t, 9, 9, nil, [][2]int{{4, 4}, {3, 3}})
	if v := resolve(g, me); !v.Absorbed() || v.Rule != rules.RuleForced {
		t.Errorf("a window with one split = %+v, want forced", v)
	}
}
