package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ozolith_test.go — #1218: the card proof for the counters-at-
// departure half of "Post-departure LKI / 'exiled with this' record"
// (docs/engine-seams.md). Every collection test destroys the creature
// with no other event in between the departure and the read, which is
// the property that was missing: MoveCard zeroes Card.Counters as
// part of the same mutation that emits the departure event, so a
// watcher reading the card AFTER the move sees none regardless of
// what it actually had.

const ozolithOracle = "1946ded1-5f53-409f-b0a6-5433bb0357d2"

// seedCreatureWithCounters puts a creature on the battlefield with
// the given counters, under owner's control.
func seedCreatureWithCounters(g *game.Game, owner uuid.UUID, counters map[string]int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Counter Bearer",
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
		Counters:   counters,
	})
	return id
}

// countersOn reads the live Counters map off a battlefield card.
func allCountersOn(t *testing.T, g *game.Game, id uuid.UUID) map[string]int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Counters
}

// --- the collection half (ability 1) --------------------------------

func TestOzolithCollectsCountersFromADepartingCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oz := seedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", ozolithOracle)
	creature := seedCreatureWithCounters(g, me.ID, map[string]int{"+1/+1": 3, "stun": 1})

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	got := allCountersOn(t, g, oz)
	if got["+1/+1"] != 3 || got["stun"] != 1 {
		t.Errorf("The Ozolith's counters = %v, want +1/+1:3 stun:1", got)
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the creature did not go to the graveyard")
	}
}

func TestOzolithIgnoresADepartureWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oz := seedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", ozolithOracle)
	creature := seedCreatureWithCounters(g, me.ID, nil)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(creature); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if got := allCountersOn(t, g, oz); len(got) != 0 {
		t.Errorf("The Ozolith's counters = %v, want none — nothing to collect", got)
	}
}

func TestOzolithIgnoresAnOpponentsCreatureLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oz := seedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", ozolithOracle)
	theirs := seedCreatureWithCounters(g, opp.ID, map[string]int{"+1/+1": 2})

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if got := allCountersOn(t, g, oz); len(got) != 0 {
		t.Errorf("The Ozolith collected an opponent's creature's counters: %v", got)
	}
}

// --- the distribution half (ability 2) ------------------------------

func TestOzolithMovesAllCountersOntoTargetCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ozID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: ozID,
		Name:       "The Ozolith",
		TypeLine:   "Legendary Artifact",
		OracleID:   ozolithOracle,
		Owner:      me.ID,
		Controller: me.ID,
		Counters:   map[string]int{"+1/+1": 2, "stun": 1},
	})
	target := seedPermanentFor(g, me.ID, "Target Bear", "Creature — Bear")

	advanceTo(t, g, game.StepBeginCombat)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)

	if got := allCountersOn(t, g, target); got["+1/+1"] != 2 || got["stun"] != 1 {
		t.Errorf("target's counters = %v, want +1/+1:2 stun:1", got)
	}
	if got := allCountersOn(t, g, ozID); len(got) != 0 {
		t.Errorf("The Ozolith kept counters after moving them all: %v", got)
	}
}

func TestOzolithSkipsTheCombatTriggerWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "The Ozolith", "Legendary Artifact", ozolithOracle)

	advanceTo(t, g, game.StepBeginCombat)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
			t.Fatalf("The Ozolith offered its combat trigger with no counters on it")
		}
	}
}
