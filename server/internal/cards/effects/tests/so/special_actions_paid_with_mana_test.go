package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// special_actions_paid_with_mana_test.go — #6247: special actions paid
// with mana. Leonin Arbiter and Quenchable Fire are their batches'; the
// shapes are registered here.
//
// CR 116.2d: "Some effects from static abilities allow a player to take
// an action to ignore the effect from that ability for a duration."
// CR 116.2c: "Some effects allow a player to take an action at a later
// time, usually … to stop a delayed triggered ability from triggering."

const (
	arbiterTestOracle = "test-6247-arbiter"
	diabolicTutor     = "14589b6b-1814-46f9-a364-83cc15dacac2"
)

func arbiterTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:                   arbiterTestOracle,
		Name:                       "Arbiter Test",
		PlayersCantSearchLibraries: true,
		IgnorableEffect:            effects.AnyPlayerMayPayToIgnore("Arbiter Test", "{2}"),
	})
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbiter := effectstest.PushCatalogPermanent(g, opp.ID, "Arbiter Test", "Creature — Cat Cleric", arbiterTestOracle, false)
	effectstest.ToMainForCost(t, g)
	return g, me, opp, arbiter
}

func searched(g *game.Game, player uuid.UUID, from int) bool {
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == player {
			return true
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == player {
			return true
		}
	}
	return false
}

// "Players can't search libraries": its controller's opponent can't, and
// neither can its controller.
func TestLeoninArbiterStopsEverySearch(t *testing.T) {
	g, me, _, _ := arbiterTable(t)
	start := len(g.Events)
	effectstest.CastCatalogSpell(t, g, "Diabolic Tutor", "Sorcery", diabolicTutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if searched(g, me.ID, start) {
		t.Fatal("a player searched a library under the Arbiter")
	}
}

// The issue's acceptance: a player pays {2} and then searches; another
// player still can't.
func TestLeoninArbiterPayTwoThenSearch(t *testing.T) {
	g, me, opp, arbiter := arbiterTable(t)
	offers := 0
	g.WithWriteLock(func() { offers = len(g.IgnoreEffectOffersLocked(me.ID)) })
	if offers != 1 {
		t.Fatalf("offers to seat 0 = %d, want the Arbiter's", offers)
	}
	effectstest.FloatForTest(g, me, "U")
	err := g.PerformSpecialAction(me.ID, arbiter, game.SpecialActionIgnoreEffect, game.SpecialActionParams{Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("ignoring for one mana = %v, want ErrInsufficientMana", err)
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.PerformSpecialAction(me.ID, arbiter, game.SpecialActionIgnoreEffect, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("paying {2} to ignore the Arbiter: %v", err)
	}
	g.WithWriteLock(func() { offers = len(g.IgnoreEffectOffersLocked(me.ID)) })
	if offers != 0 {
		t.Error("a player already ignoring the Arbiter is offered it again")
	}
	start := len(g.Events)
	effectstest.CastCatalogSpell(t, g, "Diabolic Tutor", "Sorcery", diabolicTutor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !searched(g, me.ID, start) {
		t.Fatal("the player who paid {2} should search")
	}
	// Another player has paid nothing: the Arbiter's controller still
	// can't search.
	var blocked bool
	g.WithWriteLock(func() { blocked = len(g.IgnoreEffectOffersLocked(opp.ID)) == 1 })
	if !blocked {
		t.Error("the other player should still be offered the ignore, so still bound")
	}
}

// A sacrifice is not the cost of a mana ignore.
func TestAManaIgnoreRefusesASacrifice(t *testing.T) {
	g, me, _, arbiter := arbiterTable(t)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Plains", "Basic Land — Plains")
	err := g.PerformSpecialAction(me.ID, arbiter, game.SpecialActionIgnoreEffect, game.SpecialActionParams{SacrificeID: land})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a sacrifice named for a mana ignore = %v, want ErrInvalidParam", err)
	}
}

var quenchFired int

var quenchTestBody = game.SimpleDelayedBody("test-6247/quench", func(*game.Game, *game.StackItem) error {
	quenchFired++
	return nil
})

// quenchTable schedules seat 0's "at the beginning of your next upkeep"
// trigger that seat 1 may pay {U} to stop.
func quenchTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	quenchFired = 0
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:         me.ID,
			Label:              "Quench Test — 3 more damage",
			At:                 game.StepUpkeep,
			ControllerTurnOnly: true,
			AfterThisTurn:      true,
			Body:               quenchTestBody,
			Stop: &game.DelayedTriggerStop{
				Player: opp.ID,
				Cost:   "{U}",
				Label:  "Pay {U}: Quench Test deals no additional damage to you",
			},
		})
	})
	if id == uuid.Nil {
		t.Fatal("the delayed trigger was not scheduled")
	}
	return g, me, opp, id
}

// The issue's acceptance: the target pays {U} during the same turn and
// takes no damage at the next upkeep.
func TestQuenchableFirePaidStopsTheTrigger(t *testing.T) {
	g, me, opp, id := quenchTable(t)
	var offers []game.StopDelayedTriggerOffer
	g.WithWriteLock(func() { offers = g.StopDelayedTriggerOffersLocked(opp.ID) })
	if len(offers) != 1 || offers[0].ID != id || offers[0].Cost != "{U}" {
		t.Fatalf("offers to the target = %+v, want the {U} stop", offers)
	}
	g.WithWriteLock(func() { offers = g.StopDelayedTriggerOffersLocked(me.ID) })
	if len(offers) != 0 {
		t.Fatal("only the player it names may pay to stop it")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.FloatForTest(g, opp, "U")
	if err := g.PerformSpecialAction(opp.ID, id, game.SpecialActionStopDelayedTrigger, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("paying {U} to stop it: %v", err)
	}
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.ID == id {
			t.Fatal("the paid-off trigger is still waiting")
		}
	}
	advanceToMyNextUpkeep(t, g, me.ID)
	if quenchFired != 0 {
		t.Errorf("the stopped trigger fired %d times", quenchFired)
	}
}

// Unpaid, it triggers at the next upkeep.
func TestQuenchableFireUnpaidTriggers(t *testing.T) {
	g, me, _, _ := quenchTable(t)
	advanceToMyNextUpkeep(t, g, me.ID)
	if quenchFired != 1 {
		t.Errorf("the unpaid trigger fired %d times, want 1", quenchFired)
	}
}

// advanceToMyNextUpkeep plays the table forward to `me`'s next upkeep and
// lets what triggers there resolve.
func advanceToMyNextUpkeep(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	for i := 0; i < 200; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if g.Turn.Step == game.StepUpkeep && g.Seats[g.Turn.ActiveSeat].ID == me && i > 0 {
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached the next upkeep")
}
