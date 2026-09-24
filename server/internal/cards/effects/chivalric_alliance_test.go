package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const chivalricAllianceOracle = "9b8d9236-0452-4509-aca4-a53a399fff85"

// TestChivalricAlliancePaysDiscardAndMakesAKnight is #1381: the whole
// printed second ability is "{2}, Discard a card: Create a 2/2 white
// and blue Knight creature token with vigilance." The discard is paid
// at announce (CR 602.2b) and the token only appears at resolution.
func TestChivalricAlliancePaysDiscardAndMakesAKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	alliance := pushCatalogPermanent(g, me.ID, "Chivalric Alliance", "Enchantment", chivalricAllianceOracle, false)
	discarded := ccHandCard(me, "Opt", "Instant", "{U}")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, alliance, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{discarded},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(discarded) {
		t.Error("the discard should be paid at announce")
	}
	if !me.Graveyard.Contains(discarded) {
		t.Error("the discarded card should be in the graveyard")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %d, want the {2} spent", len(me.ManaPool))
	}
	if findBattlefieldByName(g, "Knight") != uuid.Nil {
		t.Error("the Knight token should wait for resolution")
	}
	passPriorityAroundTable(t, g)
	knight := findBattlefieldByName(g, "Knight")
	if knight == uuid.Nil {
		t.Fatal("no Knight token was created")
	}
	c, ok := g.LookupCardForEffect(knight)
	if !ok {
		t.Fatal("cannot look up the Knight token")
	}
	if c.Power != 2 || c.Toughness != 2 {
		t.Errorf("Knight is %d/%d, want 2/2", c.Power, c.Toughness)
	}
	if len(c.Colors) != 2 {
		t.Errorf("Knight colors = %v, want white and blue", c.Colors)
	}
	if !game.HasKeyword(&c, "vigilance") {
		t.Error("the Knight token has no vigilance")
	}
}

// The illegal case: with no card in hand to discard, the ability
// cannot be activated at all.
func TestChivalricAllianceCannotActivateWithNoCardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	alliance := pushCatalogPermanent(g, me.ID, "Chivalric Alliance", "Enchantment", chivalricAllianceOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	// Empty the hand the demo game seeds.
	for len(me.Hand.Cards) > 0 {
		if _, err := me.Hand.PopTop(); err != nil {
			t.Fatalf("PopTop: %v", err)
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, alliance, 0, game.ActivateAbilityParams{}); err != game.ErrInvalidParam {
		t.Errorf("no card to discard: got %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 2 {
		t.Error("a refused activation must not spend its mana")
	}
}
