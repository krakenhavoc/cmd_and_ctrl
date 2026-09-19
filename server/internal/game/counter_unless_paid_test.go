package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_unless_paid_test.go — #951. A pay-unless whose decline
// counters a spell is not the pay-unless ADR 0018 §6 let the table
// walk past, and the difference is player-visible: a warded permanent
// died for free because a third seat passed priority while the payer
// had not answered.
//
// The halt is half the claim. The other half, and the one this file
// spends most of its lines on, is that the halt does not WEDGE: the
// payer can answer through it, the departure table settles it when the
// payer leaves, and it lifts by itself if the guarded spell goes away
// some other way.

// guardedSpellOnStack puts a spell controlled by `controller` on the
// stack and returns its ID.
func guardedSpellOnStack(t *testing.T, g *Game, controller *Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		pushStackSpell(t, g, Card{
			InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
			Owner: controller.ID, Controller: controller.ID,
		})
	})
	return id
}

// queueWardTax is the ward mana leg, queued straight at the engine so
// the rule is measured without a card in the way.
func queueWardTax(t *testing.T, g *Game, spell uuid.UUID, chooser uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell,
			Chooser:   chooser,
			Source:    uuid.New(),
			Cost:      "{2}",
			Question:  "Ward — pay {2} or the spell is countered",
		}); err != nil {
			t.Fatalf("QueueCounterUnlessPaidForEffect: %v", err)
		}
	})
}

func onlyPendingChoice(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	return g.PendingChoices[0]
}

// TestAWardTaxHaltsTheStack is the bug, stated. Four seats, because
// the interleaving #951 reports needs a pass from somebody other than
// the two players in the exchange — which is the ORDINARY case at a
// real table and exactly what the old ward tests never produced.
func TestAWardTaxHaltsTheStack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	queueWardTax(t, g, spell, uuid.Nil)

	prompt := onlyPendingChoice(t, g)
	if prompt.Kind != PendingChoicePayUnless {
		t.Fatalf("prompt kind = %q, want pay_unless", prompt.Kind)
	}
	// "its controller" is read off the guarded object when the caller
	// does not name a payer.
	if prompt.Chooser != caster.ID {
		t.Errorf("chooser = %s, want the spell's controller %s", prompt.Chooser, caster.ID)
	}
	if prompt.GuardsStackItem != spell {
		t.Errorf("GuardsStackItem = %s, want %s", prompt.GuardsStackItem, spell)
	}
	// It blocks WITHOUT the #567 override — the gate works the halt
	// out from the board, which is what stops the seventh card of this
	// shape being wrong again.
	if prompt.ForceBlocks {
		t.Error("the ward tax set ForceBlocks; the halt is supposed to be derived, not declared")
	}
	if !g.ChoicePromptBlocksTable(prompt) {
		t.Fatal("a pay-unless guarding a spell on the stack does not block the table")
	}
	// And the KIND is untouched: ADR 0018 §6 still holds for Rhystic
	// Study, which is the thing this fix must not undo.
	if ChoiceBlocksTable(PendingChoicePayUnless) {
		t.Error("the pay_unless KIND now blocks — #951 is per prompt, not per kind")
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
			t.Errorf("%s refusal does not name the ward tax: %v", tc.name, err)
		}
	}
	// Nothing resolved: the spell the tax is about is still there,
	// which is the whole point.
	if !g.Stack.Contains(spell) {
		t.Error("the guarded spell resolved while its ward tax was unanswered")
	}
}

// A Rhystic tax with a spell on the stack keeps ADR 0018 §6's
// latitude. The shapes are one field apart, so this is the guard
// against fixing #951 by making every pay-unless a modal lockstep.
func TestARhysticTaxStillLetsTheTablePlayOn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(caster.ID, uuid.New(), "{1}",
			"Rhystic Study — pay {1}?", func(*Game) error { return nil }); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	prompt := onlyPendingChoice(t, g)
	if prompt.GuardsStackItem != uuid.Nil {
		t.Errorf("a Rhystic tax guards %s; it should guard nothing", prompt.GuardsStackItem)
	}
	if g.ChoicePromptBlocksTable(prompt) {
		t.Fatal("an ordinary pay_unless blocks the table — ADR 0018 §6 says it must not")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority under a Rhystic tax: %v", err)
	}
	_ = spell
}

// The halt is a live read of the stack, not a flag set when the prompt
// was queued. A guarded spell that leaves some other way — countered
// underneath the trigger, fizzled — takes the halt with it, so a
// prompt can never outlive the question it is about. This is the
// anti-wedge property the derivation buys.
func TestTheHaltLiftsWhenTheGuardedSpellLeavesTheStack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	queueWardTax(t, g, spell, uuid.Nil)
	prompt := onlyPendingChoice(t, g)

	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want ErrChoicePending before the spell leaves", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
	})
	if g.ChoicePromptBlocksTable(prompt) {
		t.Error("the tax still blocks after the spell it guards left the stack")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority after the guarded spell left: %v", err)
	}
	// The prompt is still open and still answerable; it simply no
	// longer stops anybody.
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the tax still queued", len(g.PendingChoices))
	}
	if err := g.ResolvePayUnless(prompt.ID, caster.ID, false); err != nil {
		t.Errorf("the payer could not answer a tax whose spell had gone: %v", err)
	}
}

// The payer answers through the halt — the prompt's own chooser is
// never gated by it — and declining counters the spell.
func TestThePayerCanAnswerThroughTheHalt(t *testing.T) {
	for _, tc := range []struct {
		name      string
		pay       bool
		countered bool
	}{
		{"declining counters it", false, true},
		{"paying lets it resolve", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			caster := g.Seats[1]
			spell := guardedSpellOnStack(t, g, caster)
			if tc.pay {
				// Fund it: a "pay" the chooser cannot cover degrades
				// to a decline, which would measure the wrong branch.
				caster.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
			}
			queueWardTax(t, g, spell, uuid.Nil)
			prompt := onlyPendingChoice(t, g)

			if err := g.ResolvePayUnless(prompt.ID, caster.ID, tc.pay); err != nil {
				t.Fatalf("ResolvePayUnless: %v", err)
			}
			if len(g.PendingChoices) != 0 {
				t.Fatalf("pending choices = %d after the answer, want 0", len(g.PendingChoices))
			}
			if got := g.Stack.Contains(spell); got == tc.countered {
				t.Errorf("spell on the stack = %v, countered expected = %v", got, tc.countered)
			}
			// Answered, so the table moves again. A halt that survived
			// its answer would be the wedge.
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority after the tax was answered: %v", err)
			}
		})
	}
}

// Nothing is asked about an object that is not on the stack. CR 118.12
// buys off a consequence, and a consequence that can no longer happen
// is not worth a prompt — this is the check every one of the six
// counter-unless-pays cards used to spell out for itself.
func TestNothingIsAskedAboutAnObjectThatHasLeftTheStack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.WithWriteLock(func() {
		if err := g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: uuid.New(),
			Source:    uuid.New(),
			Cost:      "{2}",
			Question:  "Ward — pay {2}",
		}); err != nil {
			t.Fatalf("QueueCounterUnlessPaidForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Errorf("a spell that had already left the stack raised %d prompts", len(g.PendingChoices))
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
}

// The departure column already says what a dropped pay_unless does
// (CR 800.4f, #961: the cost is not paid, so the decline runs). #951
// only has to not break it — and the drop must free the table, since a
// halt nobody is left to answer is the wedge sprint S36 is about.
//
// The payer and the spell's controller are DIFFERENT seats here, which
// is what keeps the counter observable: CR 800.4a takes the departing
// player's own objects out of the game, so a spell owned by the payer
// would have left with them and there would be nothing to measure.
func TestAWardTaxDroppedByADepartureCountersAndFreesTheTable(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	warden, payer, controller := g.Seats[0], g.Seats[1], g.Seats[2]
	spell := guardedSpellOnStack(t, g, controller)
	// A real permanent behind the prompt, because the drop's own
	// material gate (departedChoiceObjectLocked) asks who controls the
	// object that required the choice before it runs anything.
	source := departureTestSource(g, warden.ID, "Rimeshield Frost Giant")
	g.WithWriteLock(func() {
		if err := g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell,
			Chooser:   payer.ID,
			Source:    source,
			Cost:      "{2}",
			Question:  "Ward — pay {2} or the spell is countered",
		}); err != nil {
			t.Fatalf("QueueCounterUnlessPaidForEffect: %v", err)
		}
	})

	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want ErrChoicePending while the tax is open", err)
	}
	if err := g.Concede(payer.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless {
			t.Errorf("the ward tax is still queued for %s after they left", c.Chooser)
		}
	}
	if g.Stack.Contains(spell) {
		t.Error("the departed payer did not pay, so the spell should have been countered")
	}
	if c := g.blockingChoiceLocked(); c != nil {
		t.Errorf("the table is still blocked by %q after the payer left", c.Kind)
	}
}
