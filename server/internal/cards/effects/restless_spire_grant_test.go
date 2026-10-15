package effects

import (
	"testing"

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := s58p6Push(g, me.ID, "Restless Spire", "Land", rfrbRestlessSpire, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	gl6LoseAllAbilities(g, land)
	passPriorityAroundTable(t, g)
	if !s58p6View(t, g, land).IsCreature() {
		t.Fatal("the land did not animate")
	}
	if !gl6HasKeyword(t, g, land, "first strike") {
		t.Errorf("the granted first strike was removed by an earlier removal: %v", effectiveAbilities(t, g, land))
	}
}

// A removal that begins after the activation takes the grant with it.
func TestRestlessSpireLaterRemovalTakesTheGrantedFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := s58p6Push(g, me.ID, "Restless Spire", "Land", rfrbRestlessSpire, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	rfrbAnimate(t, g, me, land, "U", "R")
	if !gl6HasKeyword(t, g, land, "first strike") {
		t.Fatal("no first strike after animating")
	}
	gl6LoseAllAbilities(g, land)
	if gl6HasKeyword(t, g, land, "first strike") {
		t.Error("a later removal left the granted first strike")
	}
}
