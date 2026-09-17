package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// choice_gate_test.go pins #730: while a prompt nobody has answered is
// in the queue, the turn-structure verbs are refused, so a prompt's
// continuation can never resume against a game that has moved on.
//
// The allowlist is the other half of the claim, and it is the half a
// future kind gets wrong by default: everything blocks EXCEPT
// pay_unless, which ADR 0018 §6 deliberately leaves loose.

// queueSacrificePromptForTest puts a creature under seat 0 and asks
// them to sacrifice it, returning the prompt. Mirrors what a
// Fleshbag Marauder does on resolution.
func queueSacrificePromptForTest(t *testing.T, g *Game) (*PendingChoice, uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	victim := pushCreatureToBattlefield(t, g, me)
	g.WithWriteLock(func() {
		if n := g.PlayerSacrificesForEffect(uuid.New(), me.ID, nil, "Test Marauder — sacrifice a creature"); n != 1 {
			t.Fatalf("queued %d sacrifice prompts, want 1", n)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	return g.PendingChoices[0], victim
}

// TestSacrificePromptBlocksAdvanceStep — the headline. The cursor
// does not move while the prompt is open, the refusal names the
// prompt, and answering it unblocks the table.
func TestSacrificePromptBlocksAdvanceStep(t *testing.T) {
	g := newActiveGame(t)
	prompt, victim := queueSacrificePromptForTest(t, g)
	before := g.Turn

	_, err := g.AdvanceStep()
	if !errors.Is(err, ErrChoicePending) {
		t.Fatalf("AdvanceStep = %v, want ErrChoicePending", err)
	}
	var cpe *ChoicePendingError
	if !errors.As(err, &cpe) {
		t.Fatalf("AdvanceStep error %T does not carry a *ChoicePendingError", err)
	}
	if cpe.ChoiceID != prompt.ID {
		t.Errorf("error names choice %s, want %s", cpe.ChoiceID, prompt.ID)
	}
	if cpe.Kind != PendingChoiceSacrifice {
		t.Errorf("error kind = %q, want %q", cpe.Kind, PendingChoiceSacrifice)
	}
	if cpe.Chooser != g.Seats[0].ID {
		t.Errorf("error chooser = %s, want %s", cpe.Chooser, g.Seats[0].ID)
	}
	if g.Turn != before {
		t.Fatalf("a refused advance_step moved the cursor to %v", g.Turn)
	}

	if err := g.ResolveSacrificeChoice(prompt.ID, g.Seats[0].ID, victim); err != nil {
		t.Fatalf("ResolveSacrificeChoice: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep after the answer: %v", err)
	}
	if g.Turn == before {
		t.Error("the cursor did not move once the prompt was answered")
	}
}

// TestSacrificePromptBlocksPassPriority — the same rule on the verb
// the bug report named first, including the pass that would wrap the
// table into a step advance.
func TestSacrificePromptBlocksPassPriority(t *testing.T) {
	g := newActiveGame(t)
	prompt, victim := queueSacrificePromptForTest(t, g)
	before := g.Turn

	// Two passes at a two-seat table is the wrap that advances the
	// step. Neither of them is allowed.
	for i := 0; i < 2; i++ {
		err := g.PassPriority()
		if !errors.Is(err, ErrChoicePending) {
			t.Fatalf("PassPriority %d = %v, want ErrChoicePending", i, err)
		}
	}
	if g.Turn != before {
		t.Fatalf("a refused pass moved the cursor to %v", g.Turn)
	}

	if err := g.ResolveSacrificeChoice(prompt.ID, g.Seats[0].ID, victim); err != nil {
		t.Fatalf("ResolveSacrificeChoice: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d after the answer: %v", i, err)
		}
	}
	if g.Turn == before {
		t.Error("two passes after the answer did not move the table on")
	}
}

// TestSacrificePromptBlocksPassTurn — pass_turn walks the cursor
// through every remaining step of the turn, so it is gated for the
// same reason and more so.
func TestSacrificePromptBlocksPassTurn(t *testing.T) {
	g := newActiveGame(t)
	prompt, victim := queueSacrificePromptForTest(t, g)
	before := g.Turn

	if err := g.PassTurn(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("PassTurn = %v, want ErrChoicePending", err)
	}
	if g.Turn != before {
		t.Fatalf("a refused pass_turn moved the cursor to %v", g.Turn)
	}
	if err := g.ResolveSacrificeChoice(prompt.ID, g.Seats[0].ID, victim); err != nil {
		t.Fatalf("ResolveSacrificeChoice: %v", err)
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn after the answer: %v", err)
	}
}

// TestPayUnlessPromptDoesNotBlockTheTable — the allowlist's only
// entry, and the reason it has one: ADR 0018 §6 accepted that play
// continues while a Rhystic Study tax is unanswered.
func TestPayUnlessPromptDoesNotBlockTheTable(t *testing.T) {
	g := newActiveGame(t)
	taxed := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(taxed.ID, uuid.New(), "{1}",
			"Rhystic Study — pay {1}?", nil); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoicePayUnless {
		t.Fatalf("queue = %+v, want one pay_unless", g.PendingChoices)
	}
	g.WithWriteLock(func() {
		if c := g.blockingChoiceLocked(); c != nil {
			t.Fatalf("pay_unless blocks the table: %+v", c)
		}
	})

	before := g.Turn
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep with a pay_unless open: %v", err)
	}
	if g.Turn == before {
		t.Error("the cursor did not move")
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("PassPriority with a pay_unless open: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Errorf("the tax prompt should still be waiting for its answer, queue = %+v", g.PendingChoices)
	}
}

// TestEveryChoiceKindBlocksUnlessAllowlisted is the deny-by-default
// claim itself. It is written against a list of kinds rather than
// against behaviour so that a kind added later — #742's choose_color
// was added while this was being written — is blocking without its
// author having to know this file exists.
func TestEveryChoiceKindBlocksUnlessAllowlisted(t *testing.T) {
	kinds := []PendingChoiceKind{
		PendingChoiceDiscardFromHand,
		PendingChoiceMana,
		PendingChoiceReplacementOrder,
		PendingChoiceOptionalReplacement,
		PendingChoiceDamageAssignment,
		PendingChoiceTriggerPrompt,
		PendingChoiceTriggerOrder,
		PendingChoicePickTarget,
		PendingChoiceSacrifice,
		PendingChoiceScry,
		PendingChoiceSurveil,
		PendingChoiceLookAtTop,
		PendingChoiceSearchLibrary,
		PendingChoiceMayCast,
		PendingChoiceChooseProtector,
		PendingChoiceLegendRule,
		PendingChoiceColor,
		PendingChoiceConfirm,
		PendingChoiceChooseCards,
		PendingChoiceEntryPayLife,
		PendingChoiceCopyTarget,
		PendingChoiceCreatureType,
		// A kind nobody has written yet. Deny by default means this
		// one blocks too.
		PendingChoiceKind("some_future_kind"),
	}
	for _, kind := range kinds {
		g := newActiveGame(t)
		g.WithWriteLock(func() {
			g.QueueChoiceForEffect(PendingChoice{
				Kind:    kind,
				Chooser: g.Seats[0].ID,
				Count:   1,
				Reason:  "probe",
			})
		})
		if _, err := g.AdvanceStep(); !errors.Is(err, ErrChoicePending) {
			t.Errorf("%s: AdvanceStep = %v, want ErrChoicePending", kind, err)
		}
		if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
			t.Errorf("%s: PassPriority = %v, want ErrChoicePending", kind, err)
		}
	}
	// And the allowlist, stated the same way.
	if _, ok := nonBlockingChoiceKinds[PendingChoicePayUnless]; !ok {
		t.Error("pay_unless is the allowlist (ADR 0018 §6)")
	}
	if len(nonBlockingChoiceKinds) != 1 {
		t.Errorf("the allowlist has %d entries; adding one is an ADR 0018 §6 amendment",
			len(nonBlockingChoiceKinds))
	}
}

// TestBlockingChoiceSkipsAllowlistedEntriesInTheQueue — a pay_unless
// ahead of a real prompt in the queue does not hide it, and the
// refusal names the blocking one rather than the first one.
func TestBlockingChoiceSkipsAllowlistedEntriesInTheQueue(t *testing.T) {
	g := newActiveGame(t)
	taxed := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(taxed.ID, uuid.New(), "{1}", "tax", nil); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	prompt, _ := func() (*PendingChoice, uuid.UUID) {
		me := g.Seats[0]
		victim := pushCreatureToBattlefield(t, g, me)
		g.WithWriteLock(func() {
			g.PlayerSacrificesForEffect(uuid.New(), me.ID, nil, "sacrifice a creature")
		})
		return g.PendingChoices[len(g.PendingChoices)-1], victim
	}()

	_, err := g.AdvanceStep()
	var cpe *ChoicePendingError
	if !errors.As(err, &cpe) {
		t.Fatalf("AdvanceStep = %v, want a *ChoicePendingError", err)
	}
	if cpe.ChoiceID != prompt.ID {
		t.Errorf("the refusal names %s (%s), want the sacrifice prompt %s",
			cpe.ChoiceID, cpe.Kind, prompt.ID)
	}
}
