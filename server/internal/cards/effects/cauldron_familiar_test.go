package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cauldronFamiliarOracle = "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e"

// pushFoodToken puts a plain Food artifact on the battlefield under
// owner's control — enough to pay a "Sacrifice a Food" cost.
func pushFoodToken(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Food", TypeLine: "Token Artifact — Food",
		Owner: owner, Controller: owner,
	})
	return id
}

// TestCauldronFamiliarReturnsFromGraveyardBySacrificingAFood is #1381:
// the whole printed second ability is "Sacrifice a Food: Return this
// card from your graveyard to the battlefield." The cost is paid at
// announce (the Food leaves before the Cat resolves), and the Cat
// returns as a new object, so its ETB drain trigger fires again — the
// Cat-Oven loop.
func TestCauldronFamiliarReturnsFromGraveyardBySacrificingAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushCatalogGraveyardCard(me, "Cauldron Familiar", "Creature — Cat", cauldronFamiliarOracle, 1, 1)
	food := pushFoodToken(g, me.ID)
	beforeMine, beforeOpp := me.Life, opp.Life

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food should be sacrificed at announce, not left behind")
	}
	if g.Battlefield.Contains(id) {
		t.Error("the Cat should be on the stack, not the battlefield, before resolution")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Cat did not return to the battlefield")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the Cat is still in the graveyard")
	}
	if me.Life != beforeMine+1 {
		t.Errorf("controller life %d, want %d — the ETB drain should fire again on return", me.Life, beforeMine+1)
	}
	if opp.Life != beforeOpp-1 {
		t.Errorf("opponent life %d, want %d — the ETB drain should fire again on return", opp.Life, beforeOpp-1)
	}
}

// The illegal case: with no Food to sacrifice, the activation is
// refused outright and the Cat stays put.
func TestCauldronFamiliarCannotActivateWithNoFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogGraveyardCard(me, "Cauldron Familiar", "Creature — Cat", cauldronFamiliarOracle, 1, 1)

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != game.ErrInvalidParam {
		t.Errorf("no Food to sacrifice: got %v, want ErrInvalidParam", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("a refused activation must leave the Cat in the graveyard")
	}

	// Naming a permanent that isn't a Food is refused the same way
	// Goblin Bombardment refuses a non-creature.
	notFood := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{notFood},
	}); err != game.ErrIllegalTarget {
		t.Errorf("sacrificing a non-Food: got %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(notFood) {
		t.Error("a rejected activation must not pay any part of its cost")
	}
}
