package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the collection half (ability 1) --------------------------------

func TestOzolithCollectsCountersFromADepartingCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oz := effectstest.SeedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", effectstest.OzolithOracle)
	creature := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{"+1/+1": 3, "stun": 1})

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	got := effectstest.AllCountersOn(t, g, oz)
	if got["+1/+1"] != 3 || got["stun"] != 1 {
		t.Errorf("The Ozolith's counters = %v, want +1/+1:3 stun:1", got)
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the creature did not go to the graveyard")
	}
}

func TestOzolithIgnoresADepartureWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oz := effectstest.SeedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", effectstest.OzolithOracle)
	creature := effectstest.SeedCreatureWithCounters(g, me.ID, nil)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.AllCountersOn(t, g, oz); len(got) != 0 {
		t.Errorf("The Ozolith's counters = %v, want none — nothing to collect", got)
	}
}

func TestOzolithIgnoresAnOpponentsCreatureLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oz := effectstest.SeedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", effectstest.OzolithOracle)
	theirs := effectstest.SeedCreatureWithCounters(g, opp.ID, map[string]int{"+1/+1": 2})

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.AllCountersOn(t, g, oz); len(got) != 0 {
		t.Errorf("The Ozolith collected an opponent's creature's counters: %v", got)
	}
}

// --- the distribution half (ability 2) ------------------------------

func TestOzolithMovesAllCountersOntoTargetCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ozID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: ozID,
		Name:       "The Ozolith",
		TypeLine:   "Legendary Artifact",
		OracleID:   effectstest.OzolithOracle,
		Owner:      me.ID,
		Controller: me.ID,
		Counters:   map[string]int{"+1/+1": 2, "stun": 1},
	})
	target := effectstest.SeedPermanentFor(g, me.ID, "Target Bear", "Creature — Bear")

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.AllCountersOn(t, g, target); got["+1/+1"] != 2 || got["stun"] != 1 {
		t.Errorf("target's counters = %v, want +1/+1:2 stun:1", got)
	}
	if got := effectstest.AllCountersOn(t, g, ozID); len(got) != 0 {
		t.Errorf("The Ozolith kept counters after moving them all: %v", got)
	}
}

func TestOzolithSkipsTheCombatTriggerWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", effectstest.OzolithOracle)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatalf("The Ozolith offered its combat trigger with no counters on it")
		}
	}
}
