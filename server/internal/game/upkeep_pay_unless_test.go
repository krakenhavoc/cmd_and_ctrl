package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// upkeep_pay_unless_test.go — #997. A pay-unless the ACTIVE player
// owes during their OWN upkeep is not the pay-unless ADR 0018 §6 let
// the table walk past, and the difference is player-visible: a table
// could take the whole turn with Stasis's {U} unpaid, or with Pact of
// Negation's {3}{U}{U} owed and the loss still pending.
//
// The halt is half the claim. The other half, and the one this file
// spends most of its lines on, is that the halt does not WEDGE: the
// payer answers through it, it lifts the moment the cursor is
// somewhere else, and the departure table settles it when the payer
// leaves.

// queueUpkeepTax queues the upkeep pay-or-else straight at the engine
// so the rule is measured without a card in the way. `declined` is
// bumped by the "or else".
func queueUpkeepTax(t *testing.T, g *Game, chooser, source uuid.UUID, declined *int) *PendingChoice {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueUpkeepPayUnlessForEffect(UpkeepPayUnlessPrompt{
			Chooser:  chooser,
			Source:   source,
			Cost:     "{U}",
			Question: "Stasis — pay {U} or sacrifice Stasis?",
			OnDecline: func(*Game) error {
				if declined != nil {
					*declined++
				}
				return nil
			},
		}); err != nil {
			t.Fatalf("QueueUpkeepPayUnlessForEffect: %v", err)
		}
	})
	return onlyPendingChoice(t, g)
}

// TestAnUpkeepPayUnlessHaltsTheTurn is the bug, stated. The cursor
// starts in seat 0's upkeep, which is where every card of this family
// asks its question.
func TestAnUpkeepPayUnlessHaltsTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("setup: the cursor is in %q, want the upkeep", g.Turn.Step)
	}
	active := g.Seats[0]
	source := departureTestSource(g, active.ID, "Stasis")
	prompt := queueUpkeepTax(t, g, active.ID, source, nil)

	if prompt.Kind != PendingChoicePayUnless {
		t.Fatalf("prompt kind = %q, want pay_unless", prompt.Kind)
	}
	// Anchored to the step it was asked in, read off the cursor. The
	// card said nothing about blocking (#997): the engine did.
	want := TurnStep{Turn: g.Turn.Seq, Step: StepUpkeep}
	if prompt.OwedInStep != want {
		t.Errorf("OwedInStep = %v, want %v", prompt.OwedInStep, want)
	}
	if !g.ChoicePromptBlocksTable(prompt) {
		t.Fatal("an upkeep pay-or-else does not block the table")
	}
	// And the KIND is untouched: ADR 0018 §6 still holds for Rhystic
	// Study, which is the thing this fix must not undo.
	if ChoiceBlocksTable(PendingChoicePayUnless) {
		t.Error("the pay_unless KIND now blocks — #997 is per prompt, not per kind")
	}

	for _, tc := range []struct {
		name string
		call func() error
	}{
		{"pass_priority", g.PassPriority},
		{"advance_step", func() error { _, err := g.AdvanceStep(); return err }},
		{"pass_turn", g.PassTurn},
	} {
		err := tc.call()
		if !errors.Is(err, ErrChoicePending) {
			t.Errorf("%s = %v, want ErrChoicePending", tc.name, err)
			continue
		}
		var cpe *ChoicePendingError
		if !errors.As(err, &cpe) || cpe.ChoiceID != prompt.ID {
			t.Errorf("%s refusal does not name the upkeep tax: %v", tc.name, err)
		}
	}
	// Nothing moved: the table is still in the upkeep the question was
	// asked in, which is the whole point (CR 500.4).
	if g.Turn.Step != StepUpkeep {
		t.Errorf("the cursor reached %q with the upkeep tax unanswered", g.Turn.Step)
	}
}

// A Rhystic tax raised in the same step keeps ADR 0018 §6's latitude.
// The two shapes are one door apart, so this is the guard against
// fixing #997 by making every pay-unless a modal lockstep.
func TestARhysticTaxRaisedInAnUpkeepStillLetsTheTablePlayOn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	payer := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(payer.ID, uuid.New(), "{1}",
			"Rhystic Study — pay {1}?", func(*Game) error { return nil }); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	prompt := onlyPendingChoice(t, g)
	if prompt.OwedInStep.NamesAStep() {
		t.Errorf("a Rhystic tax anchored itself to %v; it is owed in no step", prompt.OwedInStep)
	}
	if g.ChoicePromptBlocksTable(prompt) {
		t.Fatal("an ordinary pay_unless blocks the table — ADR 0018 §6 says it must not")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority under a Rhystic tax: %v", err)
	}
}

// Answering lifts the halt, and the "or else" is what a decline does.
// A halt that survived its own answer would be the wedge.
func TestTheUpkeepHaltLiftsWhenItIsAnswered(t *testing.T) {
	for _, tc := range []struct {
		name     string
		pay      bool
		declines int
	}{
		{"declining runs the or-else", false, 1},
		{"paying does not", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			active := g.Seats[0]
			if tc.pay {
				// Fund it: a "pay" the chooser cannot cover degrades
				// to a decline, which would measure the wrong branch.
				active.ManaPool.AddMana(ManaToken{Color: "U"})
			}
			source := departureTestSource(g, active.ID, "Stasis")
			declined := 0
			prompt := queueUpkeepTax(t, g, active.ID, source, &declined)

			if err := g.ResolvePayUnless(prompt.ID, active.ID, tc.pay); err != nil {
				t.Fatalf("ResolvePayUnless: %v", err)
			}
			if declined != tc.declines {
				t.Errorf("the or-else ran %d times, want %d", declined, tc.declines)
			}
			if len(g.PendingChoices) != 0 {
				t.Fatalf("pending choices = %d after the answer, want 0", len(g.PendingChoices))
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority after the upkeep tax was answered: %v", err)
			}
		})
	}
}

// The halt is a live comparison against the cursor, not a flag set
// when the prompt was queued. A game that reaches another step some
// other way — an undo to before the trigger, a restore, an admin
// walking the cursor by hand — is not held by a question about a step
// it has left. This is the anti-wedge property the derivation buys,
// and it is the same one #951's live stack read buys.
func TestTheUpkeepHaltLiftsWhenTheCursorLeavesTheStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[0]
	source := departureTestSource(g, active.ID, "Stasis")
	prompt := queueUpkeepTax(t, g, active.ID, source, nil)

	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want ErrChoicePending before the cursor moves", err)
	}
	// The cursor moves by some route this gate does not own.
	g.WithWriteLock(func() { g.Turn.Step = StepDraw })
	if g.ChoicePromptBlocksTable(prompt) {
		t.Error("the tax still blocks after the table left the step it was owed in")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority after the cursor left the upkeep: %v", err)
	}
	// The prompt is still open and still answerable; it simply no
	// longer stops anybody.
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the tax still queued", len(g.PendingChoices))
	}
	if err := g.ResolvePayUnless(prompt.ID, active.ID, false); err != nil {
		t.Errorf("the payer could not answer a tax whose step had passed: %v", err)
	}
}

// The next turn's upkeep is a DIFFERENT step, so an anchor that
// somehow outlived its own turn does not hold the table hostage a
// rotation later. The turn sequence is why the anchor is a TurnStep
// rather than a bare Step.
func TestTheUpkeepHaltDoesNotFollowTheGameIntoTheNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[0]
	source := departureTestSource(g, active.ID, "Stasis")
	prompt := queueUpkeepTax(t, g, active.ID, source, nil)

	g.WithWriteLock(func() { g.Turn.Seq++ })
	if g.ChoicePromptBlocksTable(prompt) {
		t.Error("a tax owed in turn 1's upkeep still blocks turn 2's")
	}
}

// The payer leaves the game still owing it. The prompt goes, and —
// the half sprint S36 is about — the table is free afterwards. A halt
// nobody is left to answer is the wedge.
//
// The "or else" does NOT run here, and that is the departure table's
// existing rule rather than this change's: the object is the departed
// player's own (Stasis is theirs, a Pact is in their graveyard), and
// CR 800.4a took it out of the game in the same breath
// (departedChoiceObjectLocked, leave_game.go). Sacrificing a permanent
// that is leaving anyway is observable, which is why that gate exists.
func TestAnUpkeepPayUnlessDroppedByADepartureFreesTheTable(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	active := g.Seats[0]
	source := departureTestSource(g, active.ID, "Stasis")
	declined := 0
	queueUpkeepTax(t, g, active.ID, source, &declined)

	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want ErrChoicePending while the tax is open", err)
	}
	if err := g.Concede(active.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			t.Errorf("the upkeep tax is still queued for %s after they left", c.Chooser)
		}
	}
	if declined != 0 {
		t.Errorf("the or-else ran %d times against a permanent CR 800.4a had already removed", declined)
	}
	if c := g.blockingChoiceLocked(); c != nil {
		t.Errorf("the table is still blocked by %q after the payer left", c.Kind)
	}
}

// A survivor's card can ask the same question of a departing seat —
// Pact of Negation after a control change, and the general shape the
// departure table's dropDecline column is written for. The cost is
// not paid (CR 800.4f), the consequence runs, and the table is free.
func TestAnUpkeepPayUnlessOnSomebodyElsesObjectDeclinesOnDeparture(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, owner.ID, "Stasis")
	declined := 0
	queueUpkeepTax(t, g, leaver.ID, source, &declined)

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if declined != 1 {
		t.Errorf("the or-else ran %d times, want 1 — CR 800.4f says the cost is not paid", declined)
	}
	if c := g.blockingChoiceLocked(); c != nil {
		t.Errorf("the table is still blocked by %q after the payer left", c.Kind)
	}
}
