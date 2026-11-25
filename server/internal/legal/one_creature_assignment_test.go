package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// one_creature_assignment_test.go — #6192, Cunning Giant: "If this creature
// is unblocked, you may have it assign its combat damage to a creature
// defending player controls." The move list offers all of it to the player
// first (the canonical answer, CR 510.1b), then all of it to each of the
// defending player's creatures, and every move is one the engine accepts.
func TestOneCreatureAssignmentMoves(t *testing.T) {
	prev := game.CatalogUnblockedAssignsToACreature
	game.CatalogUnblockedAssignsToACreature = func(key string) bool { return key == "test-cunning-giant" }
	t.Cleanup(func() { game.CatalogUnblockedAssignsToACreature = prev })

	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	giant := battlefieldCard(g, me, game.Card{Name: "Cunning Giant", TypeLine: "Creature — Giant", Power: 4, Toughness: 4, OracleID: "test-cunning-giant"})
	bear := battlefieldCard(g, def, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	wall := battlefieldCard(g, def, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 5})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)

	moves := assignmentMoves(t, g, me.ID)
	if len(moves) != 3 {
		t.Fatalf("%d moves, want the player and each of 2 creatures: %v", len(moves), labels(moves))
	}
	first, trample := splitOf(t, moves[0])
	if trample != 4 || first[bear]+first[wall] != 0 {
		t.Errorf("canonical = %v + %d to the player, want all 4 to the player", first, trample)
	}
	for _, m := range moves[1:] {
		sp, tr := splitOf(t, m)
		if tr != 0 || !((sp[bear] == 4 && sp[wall] == 0) || (sp[wall] == 4 && sp[bear] == 0)) {
			t.Errorf("move %q = %v + %d, want all 4 on one creature", m.Label, sp, tr)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}
