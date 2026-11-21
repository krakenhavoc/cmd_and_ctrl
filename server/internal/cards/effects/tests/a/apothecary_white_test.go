package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const apothecaryWhiteOracle = "7bf8bb41-6f46-492c-b6cf-5cfb29f973a4"

func TestApothecaryWhiteTapXFoodsCreatesXHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	white := effectstest.B12Push(g, me.ID, "Apothecary White", "Legendary Creature — Human Cleric", apothecaryWhiteOracle, 3, 4)
	var foods []uuid.UUID
	g.WithWriteLock(func() {
		for range 3 {
			ids, err := g.CreateTokensForEffect(me.ID, effects.FoodToken(), 1, game.TokenEntryOptions{})
			if err != nil {
				t.Fatalf("create Food: %v", err)
			}
			foods = append(foods, ids...)
		}
	})
	effectstest.AdvanceToMain(t, g)

	effectstest.B16Activate(t, g, me.ID, white, 0, game.ActivateAbilityParams{
		XValue: 2,
		TapIDs: foods[:2],
	})
	if got := effectstest.B30TokensNamed(g, me.ID, "Human"); got != 2 {
		t.Errorf("Human tokens = %d, want 2", got)
	}
	if !effectstest.B16Tapped(t, g, foods[0]) || !effectstest.B16Tapped(t, g, foods[1]) || effectstest.B16Tapped(t, g, foods[2]) {
		t.Error("the activation did not tap exactly the two Foods named")
	}
}

func TestApothecaryWhiteCreatesFoodPerPlayerBeingAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Apothecary White", "Legendary Creature — Human Cleric", apothecaryWhiteOracle, 3, 4)
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Attacker B", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("second attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B30TokensNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("Food tokens = %d, want one for each of two players attacked", got)
	}
}
