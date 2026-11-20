package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// resolution_pay_unless_test.go — #5793: a pay-unless a resolution asks
// is owed in the step it is asked in (CR 118.12, 608.2g, 117.3b), so the
// table and the resolution wait for the answer. The card-level cases
// (Crystal Shard, a test sorcery with a later instruction) are in
// cards/effects/pay_unless_resolving_spell_test.go.

// TestAPayUnlessAResolutionAsksIsOwedInItsStep: the engine derives the
// anchor; the asking ability says nothing about the step. The item stays
// on the stack, nothing passes and the step does not end until the payer
// answers, and the answer finishes the resolution.
func TestAPayUnlessAResolutionAsksIsOwedInItsStep(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	declined := false
	decline := registerTestPayUnlessDecline(t, &declined)
	item := pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "tax", func(g *Game, it *StackItem) error {
		return g.QueuePayUnlessKeyedForEffect(them.ID, uuid.Nil, "{1}", "Pay {1}?", decline, nil)
	})
	resolveUntilPrompt(t, g)

	tax := g.PendingChoices[0]
	if tax.Kind != PendingChoicePayUnless || !tax.midResolution {
		t.Fatalf("prompt kind=%q midResolution=%v, want a stamped pay_unless", tax.Kind, tax.midResolution)
	}
	if !tax.OwedInStep.IsCurrent(g.Turn) {
		t.Fatalf("the pay-unless is owed in %v, want the step it was asked in (%v)", tax.OwedInStep, g.currentTurnStepLocked())
	}
	if !g.ChoicePromptBlocksTable(tax) || !g.ResolutionPaused() {
		t.Fatalf("blocks=%v paused=%v, want both", g.ChoicePromptBlocksTable(tax), g.ResolutionPaused())
	}
	step := g.Turn.Step
	if err := g.PassPriority(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("PassPriority = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("AdvanceStep = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != step {
		t.Fatalf("the step moved from %v to %v with the pay-unless open", step, g.Turn.Step)
	}

	if err := g.ResolvePayUnless(tax.ID, them.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !declined {
		t.Error("the decline did not run")
	}
	if _, ok := g.StackMeta[item.ID]; ok {
		t.Error("the ability is still on the stack after its pay-unless was answered")
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still paused after the answer")
	}
	if g.Turn.Step != step {
		t.Errorf("the answer moved the step to %v", g.Turn.Step)
	}
}

// TestADerivedAnchorSurvivesARestore: the anchor is OwedInStep, already
// on the snapshot, so a restore taken with the question open still holds
// the table.
func TestADerivedAnchorSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	declined := false
	decline := registerTestPayUnlessDecline(t, &declined)
	pushAbilityItem(g, StackItemTriggered, me.ID, uuid.Nil, "tax", func(g *Game, it *StackItem) error {
		return g.QueuePayUnlessKeyedForEffect(them.ID, uuid.Nil, "{1}", "Pay {1}?", decline, nil)
	})
	resolveUntilPrompt(t, g)
	_, g = roundTrip(t, g)
	tax := g.PendingChoices[0]
	if !tax.OwedInStep.IsCurrent(g.Turn) || !g.ResolutionPaused() {
		t.Fatalf("restored: owed=%v paused=%v, want the anchor and the pause", tax.OwedInStep, g.ResolutionPaused())
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, ErrChoicePending) {
		t.Errorf("restored AdvanceStep = %v, want ErrChoicePending", err)
	}
}

// TestADerivedAnchorKeepsAStandingAnswer is ADR 0127's 2026-11-20
// amendment: Rhystic Study's tax is step-owed now, and still covered, so
// "Never pay" answers it and the resolution then finishes. An anchor the
// card sets itself (the upkeep pay-or-else) still leaves it uncovered.
func TestADerivedAnchorKeepsAStandingAnswer(t *testing.T) {
	g := newActiveGame(t)
	payer, drawer := g.Seats[0], g.Seats[1]
	var tax, upkeep *PendingChoice
	resolvingOpenTestRow(g, func() {
		if err := queuePayUnlessHooked(g, payer.ID, uuid.Nil, "{1}", "Test Study — pay {1}?", func(g *Game) error {
			queueConfirmHooked(g, ConfirmPrompt{Chooser: drawer.ID, Question: "Test Study — draw a card?"}, nil, nil)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		tax = lastChoice(g)
		if err := g.QueueUpkeepPayUnlessForEffect(UpkeepPayUnlessPrompt{Chooser: payer.ID, Cost: "{1}", Question: "upkeep"}); err != nil {
			t.Fatal(err)
		}
		upkeep = lastChoice(g)
	})
	if !tax.OwedInStep.IsCurrent(g.Turn) || !g.ChoicePromptBlocksTable(tax) {
		t.Fatalf("the tax: owed=%v blocks=%v, want it held in its step", tax.OwedInStep, g.ChoicePromptBlocksTable(tax))
	}
	if want := testRowKey + "pay_unless#1"; tax.AutoAnswerKey != want {
		t.Fatalf("the derived anchor took the tax's standing answer: key %q, want %q", tax.AutoAnswerKey, want)
	}
	if upkeep.AutoAnswerKey != "" {
		t.Errorf("an upkeep pay-or-else is covered (key %q); its own anchor must keep it asked", upkeep.AutoAnswerKey)
	}
	// Drop the upkeep prompt so the tax is the only one the rule meets.
	g.WithWriteLock(func() {
		for i, c := range g.PendingChoices {
			if c == upkeep {
				g.removeChoiceAtLocked(i)
				break
			}
		}
	})

	if err := g.SetAutoAnswers(payer.ID, map[string]AutoAnswer{tax.AutoAnswerKey: AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	id, chooser, ok := g.NextAutoAnswer()
	if !ok || id != tax.ID || chooser != payer.ID {
		t.Fatalf("NextAutoAnswer = %v %v %v, want the step-owed tax for its payer", id, chooser, ok)
	}
	if _, err := g.AutoAnswer(id); err != nil {
		t.Fatal(err)
	}
	if choiceByID(g, tax.ID) != nil {
		t.Fatal("the tax is still open after Never pay")
	}
	draw := lastChoice(g)
	if draw == nil || draw.Kind != PendingChoiceConfirm || draw.Chooser != drawer.ID {
		t.Fatalf("Never pay did not run the decline branch: %+v", draw)
	}
}

// resolvingOpenTestRow is resolvingTestRow with the resolution opened
// too, the way resolveTopOfStackLocked opens it, so a prompt queued in fn
// is stamped as part of it (midResolution). The resolution stays open,
// held by any blocking prompt fn queued.
func resolvingOpenTestRow(g *Game, fn func()) {
	item := &StackItem{ID: uuid.New(), Params: EffectParams{Ability: &AbilityRef{
		Key: "test-study", Slot: AbilitySlotTriggered, Ref: "own:0",
		Name: "Test Study — draw unless caster pays {1}",
	}}}
	g.WithWriteLock(func() {
		g.beginResolvingLocked(item)
		g.beginResolutionLocked()
		fn()
		g.endResolutionLocked()
		g.holdForOpenResolutionLocked()
	})
}

// registerTestPayUnlessDecline registers, once per test binary, a decline
// that flips *flag. The key is test-only.
func registerTestPayUnlessDecline(t *testing.T, flag *bool) PayUnlessElse {
	t.Helper()
	testPayUnlessDeclineFlag = flag
	return testPayUnlessDecline
}

var (
	testPayUnlessDeclineFlag *bool
	testPayUnlessDecline     = RegisterPayUnlessElse(testEffectKeyPrefix+"5793-decline", func(*Game, PayUnlessAnswer) error {
		if testPayUnlessDeclineFlag != nil {
			*testPayUnlessDeclineFlag = true
		}
		return nil
	})
)
