package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const stalactiteDaggerOracle = "064c3269-c012-4fc7-84c3-5f147a82a03c"

func TestStalactiteDaggerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(stalactiteDaggerOracle)
	if !ok || spec.Name != "Stalactite Dagger" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Entering makes the changeling Shapeshifter; equipped, a Bear is 3/3
// and every creature type.
func TestStalactiteDaggerMakesAShapeshifterAndMakesTheEquippedAllTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dagger := effectstest.RfdCast(t, g, "Stalactite Dagger", "Artifact — Equipment", stalactiteDaggerOracle, 0, 0)
	effectstest.MonSettle(t, g)
	if n := effectstest.RfdCountNamed(g, me.ID, "Shapeshifter"); n != 1 {
		t.Fatalf("%d Shapeshifter tokens, want 1", n)
	}
	shifter := effectstest.FindBattlefieldByName(g, "Shapeshifter")
	if c := effectstest.LayeredCard(t, g, shifter); !c.HasSubtype("Kithkin") || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the token is not a 1/1 changeling: %+v", c)
	}

	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if effectstest.LayeredCard(t, g, bear).HasSubtype("Elf") {
		t.Fatal("an unequipped Bear is an Elf")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dagger, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear), Strict: true}); err == nil {
		t.Fatal("equipped for {C} against Equip {2}")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, dagger, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear), Strict: true})
	if !effectstest.AttachedTo(t, g, dagger, bear) {
		t.Fatal("the Dagger is not attached")
	}
	c := effectstest.LayeredCard(t, g, bear)
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
