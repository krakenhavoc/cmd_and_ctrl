package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// suspect_test.go — #2698 (CR 701.60). A suspected creature has menace
// and can't block. The enumerator reads both through the engine's own
// block gate, so no move for the suspected blocker is offered, the plain
// blocker keeps its move, and every move that IS offered is accepted.

func TestASuspectedCreatureIsNotOfferedABlock(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	attacker := freshCreature(g, me, "Attacker")
	suspect := enteredCard(g, def, game.Card{Name: "Suspect", TypeLine: "Creature — Human", Power: 2, Toughness: 2})
	plain := enteredCard(g, def, game.Card{Name: "Plain", TypeLine: "Creature — Human", Power: 2, Toughness: 2})
	g.WithWriteLock(func() { g.SuspectForEffect(suspect) })
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	moves := legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	if got := blockTargetsOffered(t, moves, suspect); len(got) != 0 {
		t.Errorf("a suspected creature was offered blocks on %v", got)
	}
	if got := blockTargetsOffered(t, moves, plain); !got[attacker] {
		t.Errorf("the unsuspected creature beside it lost its block: %v", got)
	}
}

// A suspected ATTACKER has menace: with one untapped creature the
// defender is offered no block on it and owes no decision, and with two
// the block is offered as one grouped move the engine accepts.
func TestASuspectedAttackerHasMenaceForTheEnumerator(t *testing.T) {
	for _, blockers := range []int{1, 2} {
		g := newTable(t)
		me := g.Seats[g.Turn.ActiveSeat]
		def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(me)
		clearHand(def)
		attacker := freshCreature(g, me, "Suspect")
		for i := 0; i < blockers; i++ {
			enteredCard(g, def, game.Card{Name: "Blocker", TypeLine: "Creature — Human", Power: 1, Toughness: 4})
		}
		g.WithWriteLock(func() { g.SuspectForEffect(attacker) })
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, def.ID); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)

		moves := legal.EnumerateFor(g, def.ID)
		dispatchAll(t, g, def.ID, moves)
		singles, groups := 0, 0
		for _, m := range moves {
			switch {
			case m.Kind == legal.KindBlock && m.Type == legal.TypeDeclareBlockers:
				groups++
			case m.Kind == legal.KindBlock:
				singles++
			}
		}
		if singles != 0 {
			t.Errorf("%d blockers: %d single blocks offered on a menace attacker", blockers, singles)
		}
		if want := blockers - 1; groups != want {
			t.Errorf("%d blockers: %d grouped blocks offered, want %d", blockers, groups, want)
		}
		if owes := g.SeatOwesBlockDecision(def.ID); owes != (blockers == 2) {
			t.Errorf("%d blockers: SeatOwesBlockDecision = %v", blockers, owes)
		}
	}
}
