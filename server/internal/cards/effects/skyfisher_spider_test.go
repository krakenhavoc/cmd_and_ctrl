package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const skyfisherSpiderOracle = "67158085-4807-4f0e-8599-c55ec7ec252e"

func TestSkyfisherSpiderIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(skyfisherSpiderOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Skyfisher Spider: %+v %v", spec, ok)
	}
}

func TestSkyfisherSpiderSacrificeAnotherCreatureDestroysANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := pushTypedCard(g, opp.ID, "Their Anthem", "Enchantment", "{1}{W}{W}")
	castAndResolveCreature(t, g, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle)
	passPriorityAroundTable(t, g)
	answerChooseCardsOrPick(t, g, me.ID, fodder)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickCard(t, g, me.ID, theirs)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted nonland permanent was not destroyed")
	}
}

func TestSkyfisherSpiderDecliningTheSacrificeDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := pushTypedCard(g, opp.ID, "Their Anthem", "Enchantment", "{1}{W}{W}")
	castAndResolveCreature(t, g, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice offer")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(theirs) {
		t.Error("something left the battlefield although the sacrifice was declined")
	}
}

func TestSkyfisherSpiderDeathGainsLifePerCreatureCardAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := pushCatalogPermanent(g, me.ID, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle, false)
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	life := me.Life
	killCreature(t, g, me.ID, spider)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 (two Bears and the Spider)", life, me.Life)
	}
	if me.Graveyard.Contains(spider) {
		t.Error("the Spider is still in the graveyard")
	}
}

func TestSkyfisherSpiderDeathDeclinedKeepsTheCardAndLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := pushCatalogPermanent(g, me.ID, "Skyfisher Spider", "Creature — Spider", skyfisherSpiderOracle, false)
	life := me.Life
	killCreature(t, g, me.ID, spider)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life || !me.Graveyard.Contains(spider) {
		t.Errorf("life %d → %d, in graveyard %v: declined should change nothing", life, me.Life, me.Graveyard.Contains(spider))
	}
}
