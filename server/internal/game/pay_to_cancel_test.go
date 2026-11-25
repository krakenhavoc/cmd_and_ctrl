package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// pay_to_cancel_test.go — #6220: "that player loses 1 life at the
// beginning of their next draw step unless they pay {1} before that draw
// step" (Nafs Asp). A CR 603.7 delayed trigger at the beginning of that
// player's next draw step, with a payment the player may make any time they
// have priority before it (CR 117.1: the player with priority may take
// actions), which removes it.

var payToCancelTestLoses = DelayedBody("test/pay-to-cancel-loses-life", func(g *Game, item *StackItem, p EffectParams) error {
	return g.ChangePlayerLifeForEffect(item.SourceCardID, p.Player, -p.Amount)
})

// scheduleNafs schedules opp's "loses 1 life at the beginning of their next
// draw step unless they pay {1}" from `me`'s effect, during me's turn.
func scheduleNafs(t *testing.T, g *Game, me, opp *Player) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:  me.ID,
			Label:       "Nafs Asp — lose 1 life",
			At:          StepDraw,
			TurnOf:      opp.ID,
			Body:        payToCancelTestLoses,
			Params:      EffectParams{Player: opp.ID, Amount: 1},
			PayToCancel: &PayToCancel{Player: opp.ID, Cost: "{1}"},
		})
	})
	return id
}

// toOppDrawResolved passes through opp's next draw step, every trigger of
// it resolved, to their main phase.
func toOppDrawResolved(t *testing.T, g *Game, opp *Player) {
	t.Helper()
	for range 64 {
		if g.Turn.Step == StepPrecombatMain && g.Seats[g.Turn.ActiveSeat] == opp && g.Stack.Size() == 0 {
			return
		}
		finishPendingBlockDeclarations(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("never reached %s's main phase with an empty stack (at %s)", opp.Name, g.Turn.Step)
}

func payNafs(g *Game, opp *Player, id uuid.UUID) error {
	return g.PerformSpecialAction(opp.ID, id, SpecialActionPayToCancel, SpecialActionParams{Strict: true, AutoTap: true})
}

// Not paid: the player loses 1 life as their next draw step begins.
func TestPayToCancelUnpaidHappens(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	scheduleNafs(t, g, me, opp)
	g.WithWriteLock(func() {
		if offers := g.PayToCancelOffersLocked(opp.ID); len(offers) != 1 || offers[0].Cost != "{1}" {
			t.Errorf("offers = %+v, want the one {1} payment", offers)
		}
		if offers := g.PayToCancelOffersLocked(me.ID); len(offers) != 0 {
			t.Errorf("the other player is offered %+v", offers)
		}
	})
	before := opp.Life
	toOppDrawResolved(t, g, opp)
	if opp.Life != before-1 {
		t.Errorf("life = %d, want %d", opp.Life, before-1)
	}
}

// The acceptance: paid during the player's upkeep (any time they have
// priority before that draw step), nothing happens at the draw step.
func TestPayToCancelPaidBeforeTheStep(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	mountain := pushBattlefieldForTest(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
	id := scheduleNafs(t, g, me, opp)
	for range 64 {
		if g.Turn.Step == StepUpkeep && g.Seats[g.Turn.ActiveSeat] == opp {
			break
		}
		finishPendingBlockDeclarations(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("never reached %s's upkeep", opp.Name)
	}
	if err := payNafs(g, opp, id); err != nil {
		t.Fatalf("pay {1} in the upkeep: %v", err)
	}
	if !tappedOnBattlefield(t, g, mountain) {
		t.Error("the {1} was not paid with the Mountain")
	}
	if err := payNafs(g, opp, id); !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Errorf("a second payment = %v, want ErrSpecialActionNotOffered", err)
	}
	before := opp.Life
	toOppDrawResolved(t, g, opp)
	if opp.Life != before {
		t.Errorf("life = %d, want %d: the payment was made", opp.Life, before)
	}
}

// Only the named player may pay, only with mana they have, and not once
// the step has begun (the trigger fired and left the queue).
func TestPayToCancelRefusals(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	id := scheduleNafs(t, g, me, opp)
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := payNafs(g, me, id); !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Errorf("another player's payment = %v, want ErrSpecialActionNotOffered", err)
	}
	if err := payNafs(g, opp, id); err == nil {
		t.Error("a payment with no mana was accepted")
	}
	pushBattlefieldForTest(g, opp.ID, "Mountain", "Basic Land — Mountain", "")
	for range 64 {
		if g.Turn.Step == StepDraw && g.Seats[g.Turn.ActiveSeat] == opp {
			break
		}
		finishPendingBlockDeclarations(t, g)
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if err := payNafs(g, opp, id); !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Errorf("a payment once the draw step began = %v, want ErrSpecialActionNotOffered", err)
	}
}

// A waiting payment survives a snapshot round trip.
func TestPayToCancelSurvivesASnapshot(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	id := scheduleNafs(t, g, me, opp)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	r.WithWriteLock(func() {
		offers := r.PayToCancelOffersLocked(opp.ID)
		if len(offers) != 1 || offers[0].ID != id || offers[0].Cost != "{1}" {
			t.Errorf("restored offers = %+v", offers)
		}
	})
}
