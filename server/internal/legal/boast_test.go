package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// boast_test.go — the enumerator half of #2697 (CR 702.142). A boast
// ability that the engine refuses — the creature has not attacked, or
// it has spent every activation it may make this turn — is not a move,
// so a bot seat is never offered it (#544). The enumerator reads the
// same game.Game.BoastBlockLocked the activation path does, which is
// also what lets Birgi's second activation appear without the
// enumerator knowing she exists.

const (
	oracleFearlessLiberator = "b0b93253-6432-401e-9a01-94c41fb72c10"
	oracleBirgi             = "fb81e4d3-1d8c-4779-be62-87cf49277e51"
)

func boastTable(t *testing.T, withBirgi bool) (*game.Game, *game.Player, game.Card, func() []legal.Move) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	id := battlefieldCard(g, active, game.Card{
		Name: "Fearless Liberator", TypeLine: "Creature — Dwarf Berserker",
		ManaCost: "{1}{R}", Power: 2, Toughness: 1, OracleID: oracleFearlessLiberator,
	})
	if withBirgi {
		battlefieldCard(g, active, game.Card{
			Name: "Birgi, God of Storytelling", TypeLine: "Legendary Creature — God",
			ManaCost: "{2}{R}", Power: 3, Toughness: 3, OracleID: oracleBirgi,
		})
	}
	basics(g, active, "Mountain", 9)
	var card game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	boasts := func() []legal.Move {
		return activationsOf(legal.EnumerateFor(g, active.ID), id)
	}
	return g, active, card, boasts
}

func attackWithCard(t *testing.T, g *game.Game, active *game.Player, c game.Card) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	var defender = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.DeclareAttacker(c.InstanceID, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
}

func TestABoastAbilityIsNotEnumeratedUntilTheCreatureAttacks(t *testing.T) {
	g, active, card, boasts := boastTable(t, false)
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := boasts(); len(acts) != 0 {
		t.Fatalf("a creature that has not attacked: want no boast offered, got %v", labels(acts))
	}

	attackWithCard(t, g, active, card)
	acts := boasts()
	if len(acts) != 1 {
		t.Fatalf("after attacking: want the boast offered, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, acts)

	if err := g.ActivateCatalogAbility(active.ID, card.InstanceID, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the boast: %v", err)
	}
	if acts := boasts(); len(acts) != 0 {
		t.Errorf("a spent boast: want no activation offered, got %v — the engine refuses it", labels(acts))
	}
}

func TestBirgiKeepsTheSecondBoastEnumeratedAndNoThird(t *testing.T) {
	g, active, card, boasts := boastTable(t, true)
	attackWithCard(t, g, active, card)

	for i := 0; i < 2; i++ {
		acts := boasts()
		if len(acts) != 1 {
			t.Fatalf("boast %d under Birgi: want it offered, got %v", i+1, labels(acts))
		}
		dispatchAll(t, g, active.ID, acts)
		if err := g.ActivateCatalogAbility(active.ID, card.InstanceID, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
			t.Fatalf("activate boast %d: %v", i+1, err)
		}
	}
	if acts := boasts(); len(acts) != 0 {
		t.Errorf("a third boast under Birgi: want none offered, got %v", labels(acts))
	}
}
