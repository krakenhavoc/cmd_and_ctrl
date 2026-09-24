package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const multaniYavimayasAvatarOracle = "4b8bf64b-4800-45ff-81c6-2857f34999b5"

// TestMultaniReturnsFromGraveyardByBouncingTwoLands is #1381: the
// whole printed last ability is "{1}{G}, Return two lands you control
// to their owner's hand: Return this card from your graveyard to
// your hand." The lands are paid at announce; Multani returns to
// HAND, not the battlefield.
func TestMultaniReturnsFromGraveyardByBouncingTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogGraveyardCard(me, "Multani, Yavimaya's Avatar",
		"Legendary Creature — Elemental Avatar", multaniYavimayasAvatarOracle, 0, 0)
	a := ccLand(g, me.ID, "Forest", "Basic Land — Forest")
	b := ccLand(g, me.ID, "Forest", "Basic Land — Forest")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("both lands should go back to their owner's hand at announce")
	}
	if me.Hand.Contains(id) {
		t.Error("Multani should be on the stack, not in hand, before resolution")
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Fatal("Multani did not return to hand")
	}
	if me.Graveyard.Contains(id) {
		t.Error("Multani is still in the graveyard")
	}
	if g.Battlefield.Contains(id) {
		t.Error("Multani returns to hand, not the battlefield")
	}
}

// The illegal case: with only one land under the activator's control,
// the ability cannot be activated at all (CR 118.3 — a board that
// cannot produce the count cannot pay it).
func TestMultaniCannotActivateWithOnlyOneLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogGraveyardCard(me, "Multani, Yavimaya's Avatar",
		"Legendary Creature — Elemental Avatar", multaniYavimayasAvatarOracle, 0, 0)
	a := ccLand(g, me.ID, "Forest", "Basic Land — Forest")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{a},
	}); err != game.ErrInvalidParam {
		t.Errorf("only one land: got %v, want ErrInvalidParam", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("a refused activation must leave Multani in the graveyard")
	}
	if !g.Battlefield.Contains(a) {
		t.Error("a refused activation must not bounce the one land offered")
	}
}
