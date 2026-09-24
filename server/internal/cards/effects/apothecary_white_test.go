package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const apothecaryWhiteOracle = "7bf8bb41-6f46-492c-b6cf-5cfb29f973a4"

func TestApothecaryWhiteTapXFoodsCreatesXHumans(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	white := b12Push(g, me.ID, "Apothecary White", "Legendary Creature — Human Cleric", apothecaryWhiteOracle, 3, 4)
	var foods []uuid.UUID
	g.WithWriteLock(func() {
		for range 3 {
			ids, err := g.CreateTokensForEffect(me.ID, FoodToken(), 1, game.TokenEntryOptions{})
			if err != nil {
				t.Fatalf("create Food: %v", err)
			}
			foods = append(foods, ids...)
		}
	})
	advanceToMain(t, g)

	b16Activate(t, g, me.ID, white, 0, game.ActivateAbilityParams{
		XValue: 2,
		TapIDs: foods[:2],
	})
	if got := b30TokensNamed(g, me.ID, "Human"); got != 2 {
		t.Errorf("Human tokens = %d, want 2", got)
	}
	if !b16Tapped(t, g, foods[0]) || !b16Tapped(t, g, foods[1]) || b16Tapped(t, g, foods[2]) {
		t.Error("the activation did not tap exactly the two Foods named")
	}
}

func TestApothecaryWhiteCreatesFoodPerPlayerBeingAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Apothecary White", "Legendary Creature — Human Cleric", apothecaryWhiteOracle, 3, 4)
	a := pushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Attacker B", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("second attack: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if got := b30TokensNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("Food tokens = %d, want one for each of two players attacked", got)
	}
}
