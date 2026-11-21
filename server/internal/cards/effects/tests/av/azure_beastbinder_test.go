package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const azureBeastbinderOracle = "7bb51602-60e8-4730-b11c-2870d6b42e20"

func TestAzureBeastbinderIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(azureBeastbinderOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Azure Beastbinder: %+v %v", spec, ok)
	}
}

func TestAzureBeastbinderAttackStripsAbilitiesAndMakesACreature2x2(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := effectstest.DskB1Creature(g, me.ID, "Azure Beastbinder", "Creature — Rat Rogue", azureBeastbinderOracle, 1, 3)
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Dragon", TypeLine: "Creature — Dragon", Keywords: []string{"flying"},
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.DeclareAttack(t, g, opp.ID, binder)
	effectstest.AnswerPickTarget(t, g, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	d, _ := effectstest.BattlefieldCard(g, dragon)
	if game.HasKeyword(&d, "flying") {
		t.Error("the Dragon kept flying")
	}
	if e := d.Effective(); e.Power != 2 || e.Toughness != 2 {
		t.Errorf("the Dragon is %d/%d, want 2/2", e.Power, e.Toughness)
	}
	// Until my next turn: it lasts through the opponent's turn.
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	d, _ = effectstest.BattlefieldCard(g, dragon)
	if e := d.Effective(); e.Power != 2 || game.HasKeyword(&d, "flying") {
		t.Errorf("the effect ended early: %d/%d", e.Power, e.Toughness)
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	d, _ = effectstest.BattlefieldCard(g, dragon)
	if e := d.Effective(); e.Power != 5 || !game.HasKeyword(&d, "flying") {
		t.Errorf("the effect did not end on my next turn: %d/%d", e.Power, e.Toughness)
	}
}

func TestAzureBeastbinderLeavesANoncreatureArtifactItsSizeAndIsUpToOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := effectstest.DskB1Creature(g, me.ID, "Azure Beastbinder", "Creature — Rat Rogue", azureBeastbinderOracle, 1, 3)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Idol", TypeLine: "Artifact", Keywords: []string{"indestructible"},
		Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, binder)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, c := range pick.PickTargetCards {
		if c == mine || c == binder {
			t.Error("offered one of my own permanents")
		}
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	r, _ := effectstest.BattlefieldCard(g, rock)
	if game.HasKeyword(&r, "indestructible") {
		t.Error("the artifact kept its ability")
	}
	if r.IsCreature() {
		t.Error("the artifact became a creature")
	}
}

func TestAzureBeastbinderCantBeBlockedByPowerTwoOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := effectstest.DskB1Creature(g, me.ID, "Azure Beastbinder", "Creature — Rat Rogue", azureBeastbinderOracle, 1, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, binder)
	effectstest.AnswerPickTarget(t, g, big)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, binder); err == nil {
		t.Error("a power 2 creature blocked the Beastbinder")
	}
	if err := g.DeclareBlocker(small, binder); err != nil {
		t.Errorf("a power 1 creature could not block: %v", err)
	}
}
