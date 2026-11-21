package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const skyfisherSpiderOracle = "67158085-4807-4f0e-8599-c55ec7ec252e"

func TestSkyfisherSpiderIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(skyfisherSpiderOracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Skyfisher Spider: %+v %v", spec, ok)
	}
}

func TestSkyfisherSpiderSacrificeAnotherCreatureDestroysANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := effectstest.PushTypedCard(g, opp.ID, "Their Anthem", "Enchantment", "{1}{W}{W}")
	effectstest.CastAndResolveCreature(t, g, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, fodder)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted nonland permanent was not destroyed")
	}
}

func TestSkyfisherSpiderDecliningTheSacrificeDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := effectstest.PushTypedCard(g, opp.ID, "Their Anthem", "Enchantment", "{1}{W}{W}")
	effectstest.CastAndResolveCreature(t, g, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice offer")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(theirs) {
		t.Error("something left the battlefield although the sacrifice was declined")
	}
}

func TestSkyfisherSpiderDeathGainsLifePerCreatureCardAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.PushCatalogPermanent(g, me.ID, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle, false)
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, spider)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 (two Bears and the Spider)", life, me.Life)
	}
	if me.Graveyard.Contains(spider) {
		t.Error("the Spider is still in the graveyard")
	}
}

func TestSkyfisherSpiderDeathDeclinedKeepsTheCardAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.PushCatalogPermanent(g, me.ID, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle, false)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, spider)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || !me.Graveyard.Contains(spider) {
		t.Errorf("life %d → %d, in graveyard %v: declined should change nothing", life, me.Life, me.Graveyard.Contains(spider))
	}
}
