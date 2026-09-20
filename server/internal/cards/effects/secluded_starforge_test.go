package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const secludedStarforgeOracle = "69f55a7c-6ddf-412e-b63b-b395731a1ff2"

// Two of three abilities ship: the colorless tap and the Robot. The
// middle one is not declared at all, which is what keeps the land
// from being a two-mana pump with no price.
func TestSecludedStarforgeTapsForColorlessAndBuildsRobots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forge := b12Push(g, me.ID, "Secluded Starforge", "Land", secludedStarforgeOracle, 0, 0)
	advanceToMain(t, g)

	manas := game.ManaAbilitiesForCard(game.Card{OracleID: secludedStarforgeOracle})
	if len(manas) != 1 || manas[0].Produced != "{C}" {
		t.Fatalf("one colorless mana ability, got %+v", manas)
	}
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: secludedStarforgeOracle})
	if len(abilities) != 1 {
		t.Fatalf("exactly one CR 602 ability ships, got %d", len(abilities))
	}
	for _, a := range abilities {
		if a.Cost.TapOthers != nil {
			t.Error("the tap-X-artifacts ability is deferred, not half-declared")
		}
	}

	b16Activate(t, g, me.ID, forge, 0, game.ActivateAbilityParams{})
	robot := findBattlefieldByName(g, "Robot")
	if robot == uuid.Nil {
		t.Fatal("the {5} ability creates a Robot token")
	}
	c, ok := battlefieldCardByID(g, robot)
	if !ok {
		t.Fatal("the Robot is on the battlefield")
	}
	if c.Power != 2 || c.Toughness != 2 {
		t.Errorf("Robot is %d/%d, want 2/2", c.Power, c.Toughness)
	}
	if !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("Robot is an artifact creature: %q", c.TypeLine)
	}
	if len(c.Colors) != 0 {
		t.Errorf("Robot is colorless, got %v", c.Colors)
	}
}
