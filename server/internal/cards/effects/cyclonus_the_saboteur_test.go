package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func cyclonusRow() cards.Card {
	return transformerRow(cyclonusOracleID,
		disturbFace{"Cyclonus, the Saboteur", "Legendary Artifact Creature — Robot", "{2}{U}{B}", "2", "5"},
		disturbFace{"Cyclonus, Cybertronian Fighter", "Legendary Artifact — Vehicle", "", "5", "5"},
		"More Than Meets the Eye {5}{U}{B}\nFlying\nWhenever Cyclonus deals combat damage to a player, it connives. Then if Cyclonus's power is 5 or greater, convert it.",
		"Living metal\nFlying\nWhenever Cyclonus deals combat damage to a player, convert it. If you do, there is an additional beginning phase after this phase.",
		"Living metal", "Convert", "Flying")
}

// cyclonusConnived runs the front face's trigger: Cyclonus carries
// `counters` +1/+1 counters, the hand holds the one card named, and the
// drawn card is a land. It answers the discard with `discard` and
// returns Cyclonus's state afterwards.
func cyclonusConnived(t *testing.T, counters int, discardNonland bool) (*game.Game, uuid.UUID, game.Card) {
	t.Helper()
	g, me, opp := flipMain(t)
	id := transformerIn(t, g, cyclonusRow(), me)
	if counters > 0 {
		pfAddCounters(g, id, "+1/+1", counters)
	}
	advanceTo(t, g, game.StepBeginCombat)
	tfSettle(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardForTest(me, "Spare Spell", "Sorcery", "")
	land := handCardForTest(me, "Spare Land", "Basic Land — Forest", "")
	emitCombatDamage(g, id, opp.ID, 2)
	tfSettle(t, g)
	choice := spell
	if !discardNonland {
		choice = land
	}
	answerDiscard(t, g, me.ID, choice)
	tfSettle(t, g)
	return g, id, flipState(t, g, id)
}

// Connive: discarding a nonland card puts a counter on Cyclonus, and
// with its power then 5 or greater it converts.
func TestCyclonusConvertsWhenConniveBringsItToFivePower(t *testing.T) {
	g, id, c := cyclonusConnived(t, 2, true)
	if c.Counters["+1/+1"] != 3 {
		t.Fatalf("%d counters, want 3 from the nonland discard", c.Counters["+1/+1"])
	}
	wantFace(t, g, id, 1, "Cyclonus, Cybertronian Fighter")
}

// A land discard adds no counter, so power stays under 5 and it stays
// put.
func TestCyclonusStaysWhenConniveStopsShortOfFive(t *testing.T) {
	g, id, c := cyclonusConnived(t, 2, false)
	if c.Counters["+1/+1"] != 2 {
		t.Fatalf("%d counters, want 2 (a land discard adds none)", c.Counters["+1/+1"])
	}
	wantFace(t, g, id, 0, "Cyclonus, the Saboteur")
	// CR 701.50f, #4641: it connived once, and said so.
	if evs := eventsOfKind(g, game.EventConnived); len(evs) != 1 || evs[0].CardID != id {
		t.Errorf("connive events = %+v, want one for Cyclonus", evs)
	}
}

// Cyclonus with power already 5 converts after any connive, even a land
// discard ("Then if Cyclonus's power is 5 or greater").
func TestCyclonusConvertsOnAnyConniveWhenPowerIsAlreadyFive(t *testing.T) {
	g, id, _ := cyclonusConnived(t, 3, false)
	wantFace(t, g, id, 1, "Cyclonus, Cybertronian Fighter")
}

// The Vehicle converts back on combat damage and adds a beginning phase
// after the combat phase.
func TestCyclonusCybertronianFighterAddsABeginningPhase(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, cyclonusRow(), me)
	advanceTo(t, g, game.StepBeginCombat)
	tfSettle(t, g)
	emitCombatDamage(g, id, opp.ID, 5)
	tfSettle(t, g)
	wantFace(t, g, id, 0, "Cyclonus, the Saboteur")

	seq := g.Turn.Seq
	sawUpkeep := false
	for i := 0; i < 20 && !sawUpkeep && g.Turn.Seq == seq; i++ {
		next := stepOnce(t, g)
		if next.Seq == seq && next.Step == game.StepUpkeep {
			sawUpkeep = true
		}
	}
	if !sawUpkeep {
		t.Error("no additional beginning phase (upkeep) after the combat phase in the same turn")
	}
}

// CR 701.28e: a Cyclonus that turned over since the trigger went on the
// stack neither converts nor adds a phase ("if you do").
func TestCyclonusNoPhaseWhenTheConvertIsIgnored(t *testing.T) {
	g, me, opp := flipMain(t)
	id := transformerConverted(t, g, cyclonusRow(), me)
	advanceTo(t, g, game.StepBeginCombat)
	tfSettle(t, g)
	emitCombatDamage(g, id, opp.ID, 5)
	// Let the trigger go on the stack, then turn it over in response.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.ConvertPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	tfSettle(t, g)
	c := flipState(t, g, id)
	if c.TurnOvers != 2 {
		t.Fatalf("turn-overs = %d, want the trigger's convert ignored (2)", c.TurnOvers)
	}
	seq := g.Turn.Seq
	for i := 0; i < 20 && g.Turn.Seq == seq; i++ {
		next := stepOnce(t, g)
		if next.Seq == seq && next.Step == game.StepUpkeep {
			t.Fatal("an additional beginning phase was added although the convert did nothing")
		}
	}
}
