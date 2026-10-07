package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// source_relative_targets_test.go — #2146. A target "with power less
// than this creature's power" is one set everywhere: the enumerator the
// bot picks from, the legal_targets the client's picker highlights, and
// the engine's announce check all read TargetSpec.SourceOK through the
// same walk.

const oracleUnlivingPsychopath = "038463c8-d2e1-4e7b-829d-67a744ae2660"

func TestUnlivingPsychopathEnumeratorAndViewAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	psycho := battlefieldCard(g, active, game.Card{
		Name: "Unliving Psychopath", TypeLine: "Creature — Zombie Assassin", OracleID: oracleUnlivingPsychopath,
		Power: 0, Toughness: 4, Counters: map[string]int{game.CounterPlusOne: 2},
	})
	one := battlefieldCard(g, opp, creature("One", "{G}", 1, 1))
	two := battlefieldCard(g, opp, creature("Two", "{1}{G}", 2, 2))
	three := battlefieldCard(g, opp, creature("Three", "{2}{G}", 3, 3))
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	// The Psychopath is a 2-power creature itself, so "lesser" excludes
	// it and every creature at power 2 or more; only the 1-power bear
	// is a legal destroy target.
	seen := boundedTargetsOffered(t, moves, psycho, map[string]bool{one.String(): true})
	if !seen[one.String()] {
		t.Errorf("the 1-power creature was never offered: %v", labels(moves))
	}
	if seen[two.String()] || seen[three.String()] || seen[psycho.String()] {
		t.Errorf("a creature that is not lesser was offered: %v", seen)
	}
	dispatchAll(t, g, active.ID, moves)

	// The wire's legal_targets for the destroy ability is the same set.
	v := protocol.ViewOfGameFor(g, active.ID.String())
	var cards []string
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != psycho.String() {
			continue
		}
		for _, ab := range c.ActivatedAbilities {
			if ab.LegalTargets != nil && len(ab.LegalTargets.Cards) > 0 {
				cards = ab.LegalTargets.Cards
			}
		}
	}
	if len(cards) != 1 || cards[0] != one.String() {
		t.Errorf("view legal_targets = %v, want only %s", cards, one)
	}

	// Shrink the source: the same walk now offers nothing.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == psycho {
			g.Battlefield.Cards[i].Counters = nil
		}
	}
	for _, m := range activationsOf(legal.EnumerateFor(g, active.ID), psycho) {
		if len(boundTargetIDs(t, m)) > 0 {
			t.Errorf("%q still offers targets once the source's power is 0", m.Label)
		}
	}
}
