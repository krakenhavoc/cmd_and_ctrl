package game

import (
	"testing"

	"github.com/google/uuid"
)

// leave_game_pay_unless_test.go — #961 / CR 800.4f, the rule one
// number before the reassignment ADR 0060's amendment implements:
//
//	800.4f  If an object requires a player who has left the game to pay
//	        a cost or choose whether to pay a cost, that cost is not
//	        paid.
//
// Not paying is an ANSWER, not silence: "unless that player pays {1}"
// is a sentence about what happens when the payment does not, so the
// drop of a pay_unless runs the prompt's decline branch. The table's
// drop-action column (choiceDepartureDecisions, leave_game.go) is where
// that is declared; the elimination sweep executes it.
//
// Four seats, because CR 800.4's interesting cases are unobservable in
// a two-player game — the departure that leaves one player standing
// ends it.

// queueDepartureTax queues the pay-unless a leaving player owes an
// object somebody else controls, with `then` as its decline branch.
func queueDepartureTax(t *testing.T, g *Game, chooser, source uuid.UUID, then func(*Game) error) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(chooser, source, "{1}",
			"Rhystic Study — pay {1}?", then); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
}

// TestDroppedPayUnlessRunsItsDeclineConsequence is the headline.
// Nobody inherits a departed player's Rhystic tax — that is 800.4f's
// first half, pinned by TestPayUnlessIsNeverReassigned — but the cost
// is not paid, and the Study's controller draws for it.
func TestDroppedPayUnlessRunsItsDeclineConsequence(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	taxer, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, taxer.ID, "Rhystic Study")
	hand := taxer.Hand.Size()

	declines := 0
	queueDepartureTax(t, g, leaver.ID, sourceID, func(g *Game) error {
		declines++
		return g.drawCardLocked(taxer.ID)
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if declines != 1 {
		t.Errorf("the decline branch ran %d times, want exactly 1", declines)
	}
	if got := taxer.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn off the dropped tax = %d, want 1", got)
	}
	// The prompt is still GONE, and the drop is still announced: the
	// consequence is what the drop DOES, not a reason to keep asking.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			t.Errorf("the tax is still queued, for %s", c.Chooser)
		}
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, leaver.ID) == nil {
		t.Errorf("no EventPendingChoiceDropped for the settled tax")
	}
}

// A tax the departed player owed their OWN permanent runs nothing.
// Cumulative upkeep is that shape: "sacrifice this permanent unless
// you pay its upkeep cost" is asked of the permanent's own controller,
// and CR 800.4a takes that permanent out of the game in the same
// breath as the departure. Sacrificing it instead would be observable
// — a sacrifice is a death, and other players' triggers watch for one.
func TestDroppedPayUnlessOnTheirOwnObjectRunsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	sourceID := departureTestSource(g, leaver.ID, "Mystic Remora")

	ran := false
	g.WithWriteLock(func() {
		if err := g.QueueUpkeepPayUnlessForEffect(UpkeepPayUnlessPrompt{
			Chooser:  leaver.ID,
			Source:   sourceID,
			Cost:     "{1}",
			Question: "Mystic Remora — cumulative upkeep {1}",
			OnDecline: func(*Game) error {
				ran = true
				return nil
			},
		}); err != nil {
			t.Fatalf("QueueUpkeepPayUnlessForEffect: %v", err)
		}
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if ran {
		t.Errorf("the cumulative upkeep of a permanent that left the game was resolved anyway")
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, leaver.ID) == nil {
		t.Errorf("no EventPendingChoiceDropped for the dropped own-object tax")
	}
}

// The continuation runs INSIDE the elimination sweep, which is the
// careful class of change (#808). A prompt it queues to the seat that
// just left must not survive — QueueChoiceForEffect's own guard (#864)
// refuses one — and a prompt it queues to a survivor must, because
// that seat really does owe it.
func TestTheDeclineConsequenceNeverQueuesToTheDepartedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	taxer, leaver, survivor := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, taxer.ID, "Rhystic Study")

	queueDepartureTax(t, g, leaver.ID, sourceID, func(g *Game) error {
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser: leaver.ID, Source: sourceID, Question: "asked of the seat that left",
		})
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser: survivor.ID, Source: sourceID, Question: "asked of a seat still here",
		})
		return nil
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	var queued []uuid.UUID
	for _, c := range g.PendingChoices {
		if c != nil {
			queued = append(queued, c.Chooser)
		}
	}
	if len(queued) != 1 || queued[0] != survivor.ID {
		t.Fatalf("prompts left after the sweep = %v, want exactly one, for the survivor %s",
			queued, survivor.ID)
	}
}

// The drop action is engine state like any other: rewinding to before
// the concede puts the seat, the prompt and its frame back and takes
// the drawn card back with them, and replaying the departure draws it
// again.
func TestUndoAcrossADroppedPayUnlessReplaysTheConsequence(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	taxer, leaver := g.Seats[0], g.Seats[1]
	sourceID := departureTestSource(g, taxer.ID, "Rhystic Study")
	hand := taxer.Hand.Size()
	queueDepartureTax(t, g, leaver.ID, sourceID, func(g *Game) error {
		return g.drawCardLocked(taxer.ID)
	})

	beforeConcede := g.Clone()
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if got := g.Seats[0].Hand.Size() - hand; got != 1 {
		t.Fatalf("cards drawn off the dropped tax = %d, want 1", got)
	}

	// The seat pointers are the restored game's from here on.
	g.WithWriteLock(func() { g.RestoreFrom(beforeConcede) })
	if g.Seats[1].Eliminated {
		t.Fatalf("the rewind left the conceding seat eliminated")
	}
	if got := g.Seats[0].Hand.Size(); got != hand {
		t.Fatalf("hand size after the rewind = %d, want the pre-concede %d", got, hand)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoicePayUnless {
		t.Fatalf("the rewind did not put the tax back: %+v", g.PendingChoices)
	}

	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede (replayed): %v", err)
	}
	if got := g.Seats[0].Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn off the replayed departure = %d, want the same 1", got)
	}
}

// TestCrossTableConfirmIsReassignedWhenItsAskerSurvives — the sibling
// field. A confirm queued with FromPlayer defaulted is the chooser's
// own question and is dropped; one a card addresses ACROSS the table
// is an object's choice that is not a cost, so CR 800.4g moves it. No
// card sets FromPlayer today, which is why the prompt is built here
// the way such a card would build it — the same precedent
// TestPickTargetIsReassignedAndItsTargetSetPruned is written on.
func TestCrossTableConfirmIsReassignedWhenItsAskerSurvives(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	asker, leaver, next := g.Seats[0], g.Seats[1], g.Seats[2]
	sourceID := departureTestSource(g, asker.ID, "Combustible Gearhulk")

	var own, cross uuid.UUID
	g.WithWriteLock(func() {
		own = g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser: leaver.ID, Source: sourceID, Question: "a question about your own cards",
		})
		cross = g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser: leaver.ID, FromPlayer: asker.ID, Source: sourceID,
			Question: "a question about the asker's cards",
		})
	})
	if c := findChoice(g, own); c == nil || c.FromPlayer != leaver.ID {
		t.Fatalf("setup: a defaulted FromPlayer must be the chooser: %+v", c)
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, own); c != nil {
		t.Errorf("an own-material confirm was reassigned to %s", c.Chooser)
	}
	c := findChoice(g, cross)
	if c == nil || c.Chooser != next.ID {
		t.Fatalf("the cross-table confirm went to %+v, want seat %s", c, next.ID)
	}
}
