package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const azureBeastbinderOracle = "7bb51602-60e8-4730-b11c-2870d6b42e20"

func TestAzureBeastbinderIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(azureBeastbinderOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Azure Beastbinder: %+v %v", spec, ok)
	}
}

func TestAzureBeastbinderAttackStripsAbilitiesAndMakesACreature2x2(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := dskB1Creature(g, me.ID, "Azure Beastbinder", "Creature — Rat Rogue", azureBeastbinderOracle, 1, 3)
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Dragon", TypeLine: "Creature — Dragon", Keywords: []string{"flying"},
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	declareAttack(t, g, opp.ID, binder)
	answerPickTarget(t, g, dragon)
	passPriorityAroundTable(t, g)
	d, _ := battlefieldCard(g, dragon)
	if game.HasKeyword(&d, "flying") {
		t.Error("the Dragon kept flying")
	}
	if e := d.Effective(); e.Power != 2 || e.Toughness != 2 {
		t.Errorf("the Dragon is %d/%d, want 2/2", e.Power, e.Toughness)
	}
	// Until my next turn: it lasts through the opponent's turn.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	d, _ = battlefieldCard(g, dragon)
	if e := d.Effective(); e.Power != 2 || game.HasKeyword(&d, "flying") {
		t.Errorf("the effect ended early: %d/%d", e.Power, e.Toughness)
	}
	advanceToStepOf(t, g, 0, game.StepUpkeep)
	d, _ = battlefieldCard(g, dragon)
	if e := d.Effective(); e.Power != 5 || !game.HasKeyword(&d, "flying") {
		t.Errorf("the effect did not end on my next turn: %d/%d", e.Power, e.Toughness)
	}
}

func TestAzureBeastbinderLeavesANoncreatureArtifactItsSizeAndIsUpToOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := dskB1Creature(g, me.ID, "Azure Beastbinder", "Creature — Rat Rogue", azureBeastbinderOracle, 1, 3)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Idol", TypeLine: "Artifact", Keywords: []string{"indestructible"},
		Owner: opp.ID, Controller: opp.ID,
	})
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, binder)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, c := range pick.PickTargetCards {
		if c == mine || c == binder {
			t.Error("offered one of my own permanents")
		}
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	r, _ := battlefieldCard(g, rock)
	if game.HasKeyword(&r, "indestructible") {
		t.Error("the artifact kept its ability")
	}
	if r.IsCreature() {
		t.Error("the artifact became a creature")
	}
}

func TestAzureBeastbinderCantBeBlockedByPowerTwoOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	binder := dskB1Creature(g, me.ID, "Azure Beastbinder", "Creature — Rat Rogue", azureBeastbinderOracle, 1, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	declareAttack(t, g, opp.ID, binder)
	answerPickTarget(t, g, big)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, binder); err == nil {
		t.Error("a power 2 creature blocked the Beastbinder")
	}
	if err := g.DeclareBlocker(small, binder); err != nil {
		t.Errorf("a power 1 creature could not block: %v", err)
	}
}
