package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theShireOracle = "9abf9a0e-8e7d-406b-a01d-d4870b30134e"

// TestTheShireTapsAnotherCreatureToMakeAFood is #1381: the whole
// printed third ability is "{1}{G}, {T}, Tap an untapped creature you
// control: Create a Food token." Both the land and the named creature
// tap to pay the cost, and the token waits for resolution.
func TestTheShireTapsAnotherCreatureToMakeAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shire := pushCatalogPermanent(g, me.ID, "The Shire", "Legendary Land", theShireOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, shire, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, ok := g.LookupCardForEffect(shire); !ok || !c.Tapped {
		t.Error("The Shire must tap for its own {T} component")
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Tapped {
		t.Error("the named creature must be tapped to pay the cost")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %d, want the {1}{G} spent", len(me.ManaPool))
	}
	if findBattlefieldByName(g, "Food") != uuid.Nil {
		t.Error("the Food token should wait for resolution")
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Food") == uuid.Nil {
		t.Fatal("no Food token was created")
	}
}

// The illegal cases: CR 702.184a's "another" — The Shire cannot tap
// itself to pay the creature component — and with no other creature
// on the board at all, the ability cannot be activated.
func TestTheShireCannotTapItselfOrActivateWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shire := pushCatalogPermanent(g, me.ID, "The Shire", "Legendary Land", theShireOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, shire, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{shire},
	}); err != game.ErrInvalidParam {
		t.Errorf("naming The Shire itself: got %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, shire, 0, game.ActivateAbilityParams{}); err != game.ErrInvalidParam {
		t.Errorf("no creature named: got %v, want ErrInvalidParam", err)
	}
	if c, ok := g.LookupCardForEffect(shire); !ok || c.Tapped {
		t.Error("a refused activation must not tap The Shire")
	}
	if len(me.ManaPool) != 2 {
		t.Error("a refused activation must not spend its mana")
	}
}
