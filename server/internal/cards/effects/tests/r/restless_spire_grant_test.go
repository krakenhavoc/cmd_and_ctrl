package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restless_spire_grant_test.go — #3808. The Spire's quoted "During your
// turn, this creature has first strike" is a granted ability (ADR 0093),
// so ability removal sorts against it by timestamp (CR 613.7a).

// Humility enters in response to the activation: the removal began before
// the animation resolves, so it leaves the first strike the animation
// grants later (CR 613.6, 613.7a); the land's own printed ability would
// have been removed along with the rest.
func TestRestlessSpireGrantedFirstStrikeSurvivesAnEarlierRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.S58p6Push(g, me.ID, "Restless Spire", "Land", effectstest.RfrbRestlessSpire, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.Gl6LoseAllAbilities(g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.S58p6View(t, g, land).IsCreature() {
		t.Fatal("the land did not animate")
	}
	if !effectstest.Gl6HasKeyword(t, g, land, "first strike") {
		t.Errorf("the granted first strike was removed by an earlier removal: %v", effectstest.EffectiveAbilities(t, g, land))
	}
}

// A removal that begins after the activation takes the grant with it.
func TestRestlessSpireLaterRemovalTakesTheGrantedFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.S58p6Push(g, me.ID, "Restless Spire", "Land", effectstest.RfrbRestlessSpire, "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	rfrbAnimate(t, g, me, land, "U", "R")
	if !effectstest.Gl6HasKeyword(t, g, land, "first strike") {
		t.Fatal("no first strike after animating")
	}
	effectstest.Gl6LoseAllAbilities(g, land)
	if effectstest.Gl6HasKeyword(t, g, land, "first strike") {
		t.Error("a later removal left the granted first strike")
	}
}
