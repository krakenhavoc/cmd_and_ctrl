package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const stalactiteDaggerOracle = "064c3269-c012-4fc7-84c3-5f147a82a03c"

func TestStalactiteDaggerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(stalactiteDaggerOracle)
	if !ok || spec.Name != "Stalactite Dagger" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Entering makes the changeling Shapeshifter; equipped, a Bear is 3/3
// and every creature type.
func TestStalactiteDaggerMakesAShapeshifterAndMakesTheEquippedAllTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dagger := rfdCast(t, g, "Stalactite Dagger", "Artifact — Equipment", stalactiteDaggerOracle, 0, 0)
	monSettle(t, g)
	if n := rfdCountNamed(g, me.ID, "Shapeshifter"); n != 1 {
		t.Fatalf("%d Shapeshifter tokens, want 1", n)
	}
	shifter := findBattlefieldByName(g, "Shapeshifter")
	if c := layeredCard(t, g, shifter); !c.HasSubtype("Kithkin") || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the token is not a 1/1 changeling: %+v", c)
	}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if layeredCard(t, g, bear).HasSubtype("Elf") {
		t.Fatal("an unequipped Bear is an Elf")
	}
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dagger, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear), Strict: true}); err == nil {
		t.Fatal("equipped for {C} against Equip {2}")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, dagger, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear), Strict: true})
	if !attachedTo(t, g, dagger, bear) {
		t.Fatal("the Dagger is not attached")
	}
	c := layeredCard(t, g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("equipped Bear is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	for _, tribe := range []string{"Elf", "Goblin", "Bear", "Kithkin"} {
		if !c.HasSubtype(tribe) {
			t.Errorf("the equipped Bear is not a %s", tribe)
		}
	}
	if c.HasSubtype("Equipment") {
		t.Error("all creature types made the Bear an Equipment")
	}
}
