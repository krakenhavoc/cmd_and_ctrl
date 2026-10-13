package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// zone_grant_test.go — the enumerator half of #3233 (ADR 0093's
// amendment of 2026-10-13). A bot picks from this list and from nothing
// else, so a hand card's GRANTED ability has to be in it exactly when
// the engine would accept the activation: the same row reader feeds the
// activation path, the wire and this walk.

const (
	oracleSatoruUmezawa = "ed1d21ba-ddfe-4f82-9e37-93f9e2c7ba33"
	oracleVanillaFixure = "zone-grant-enumerator-fixture"
)

// With Satoru Umezawa out, a creature card in the hand is offered the
// granted ninjutsu once blockers are declared, a noncreature card never
// is, and every move offered is one the dispatcher accepts. Without
// Satoru the same hand has no ninjutsu move at all.
func TestAGrantedNinjutsuIsEnumeratedForABot(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island2", "Island"))
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, basic("Swamp2", "Swamp"))
	attacker := battlefieldCard(g, active, vanillaAttacker("Sneaky Rat"))
	ninja := handCard(active, game.Card{
		Name: "Plain Ninja", TypeLine: "Creature — Human Ninja", Power: 2, Toughness: 2, OracleID: oracleVanillaFixure,
	})
	sorcery := handCard(active, game.Card{Name: "Plain Sorcery", TypeLine: "Sorcery", OracleID: oracleVanillaFixure})

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), ninja); len(acts) != 0 {
		t.Fatalf("no Satoru, yet the hand card is offered %v", labels(acts))
	}

	satoru := battlefieldCard(g, active, game.Card{
		Name: "Satoru Umezawa", TypeLine: "Legendary Creature — Human Ninja", Power: 2, Toughness: 4,
		OracleID: oracleSatoruUmezawa,
	})
	g.BumpLayerVersionForTest()

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, ninja)
	if len(acts) != 1 {
		t.Fatalf("want exactly one granted ninjutsu move, got %v", labels(acts))
	}
	if acts := activationsOf(moves, sorcery); len(acts) != 0 {
		t.Errorf("a sorcery card is offered %v, want nothing", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)

	// Satoru leaves: the move goes with him.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(satoru); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), ninja); len(acts) != 0 {
		t.Errorf("Satoru is gone, yet the hand card is offered %v", labels(acts))
	}
}
