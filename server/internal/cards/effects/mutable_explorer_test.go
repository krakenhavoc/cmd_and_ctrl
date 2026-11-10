package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const mutableExplorerOracle = "34b93a66-de1e-4133-a479-e0d90ba94a3a"

func TestMutableExplorerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(mutableExplorerOracle)
	if !ok || spec.Name != "Mutable Explorer" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Entering makes a tapped Mutavault token that is the card Mutavault:
// it taps for {C} and animates into a 2/2 with every creature type.
func TestMutableExplorerMakesATappedMutavaultThatWorks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	explorer := rfdCast(t, g, "Mutable Explorer", "Creature — Shapeshifter", mutableExplorerOracle, 1, 1)
	monSettle(t, g)
	if !layeredCard(t, g, explorer).HasSubtype("Kithkin") {
		t.Error("Mutable Explorer is not every creature type")
	}
	if n := rfdCountNamed(g, me.ID, "Mutavault"); n != 1 {
		t.Fatalf("%d Mutavault tokens, want 1", n)
	}
	vault := findBattlefieldByName(g, "Mutavault")
	c := layeredCard(t, g, vault)
	if !c.IsToken() || !c.IsLand() || c.IsCreature() || len(c.Colors) != 0 {
		t.Errorf("the token is not a colourless land token: %+v", c)
	}
	if !isTapped(g, vault) {
		t.Error("the Mutavault token entered untapped")
	}
	if err := g.ActivateManaAbility(me.ID, vault, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a tapped Mutavault token tapped for mana")
	}

	g.WithWriteLock(func() { untapForTest(g, vault) })
	before := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, vault, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if me.ManaPool.Total() != before+1 {
		t.Fatalf("pool %d -> %d, want one more", before, me.ManaPool.Total())
	}

	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate with the {C}: %v", err)
	}
	passPriorityAroundTable(t, g)
	c = layeredCard(t, g, vault)
	if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.HasSubtype("Elf") {
		t.Errorf("animated token: creature %v land %v %d/%d Elf %v, want a 2/2 land creature Elf",
			c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness(), c.HasSubtype("Elf"))
	}
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("animated again from an empty pool")
	}
}
