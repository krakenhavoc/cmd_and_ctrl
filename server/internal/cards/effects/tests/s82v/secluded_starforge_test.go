package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const secludedStarforgeOracle = "69f55a7c-6ddf-412e-b63b-b395731a1ff2"

func TestSecludedStarforgeTapXArtifactsPumpsByTheAnnouncedCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forge := effectstest.B12Push(g, me.ID, "Secluded Starforge", "Land", secludedStarforgeOracle, 0, 0)
	one := effectstest.B12Permanent(g, me.ID, "Rock One", "Artifact")
	two := effectstest.B12Permanent(g, me.ID, "Rock Two", "Artifact")
	target := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: secludedStarforgeOracle})
	if len(abilities) != 2 || !game.TapOthersCountFromX(abilities[0].Cost.TapOthers) || !abilities[0].Cost.DemandsX() {
		t.Fatalf("tap-X and Robot abilities = %+v", abilities)
	}

	effectstest.B16Activate(t, g, me.ID, forge, 0, game.ActivateAbilityParams{
		XValue:  2,
		TapIDs:  []uuid.UUID{one, two},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if !effectstest.B16Tapped(t, g, one) || !effectstest.B16Tapped(t, g, two) {
		t.Error("the two artifacts were not tapped as the cost")
	}
	if got := effectstest.EffectivePower(t, g, target); got != 4 {
		t.Errorf("target power = %d, want 4 after +2/+0", got)
	}
}

func TestSecludedStarforgeTapsForColorlessAndBuildsRobots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forge := effectstest.B12Push(g, me.ID, "Secluded Starforge", "Land", secludedStarforgeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)

	manas := game.ManaAbilitiesForCard(game.Card{OracleID: secludedStarforgeOracle})
	if len(manas) != 1 || manas[0].Produced != "{C}" {
		t.Fatalf("one colorless mana ability, got %+v", manas)
	}

	effectstest.B16Activate(t, g, me.ID, forge, 1, game.ActivateAbilityParams{})
	robot := effectstest.FindBattlefieldByName(g, "Robot")
	if robot == uuid.Nil {
		t.Fatal("the {5} ability creates a Robot token")
	}
	c, ok := effectstest.BattlefieldCardByID(g, robot)
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
