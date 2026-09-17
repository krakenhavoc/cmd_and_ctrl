package game

import (
	"testing"

	"github.com/google/uuid"
)

// optional_in_order_test.go pins #847: the multi-effect CR 616 order
// path asks a CR 614.10 "may" its own question, and an order nobody
// chose never fires one at all.
//
// Two gaps, one shape. The single-effect branch of
// applyReplacementsLocked has always offered an Optional replacement
// to its controller before firing it; the paths that apply SEVERAL
// effects did not:
//
//   - ResolveReplacementOrder's chosen-order loop fired the Replace
//     inline, so a "may" that happened to share a window with any
//     other effect decided itself, in the direction that favours it —
//     the same bug the CopySelector and EntryLifeCost branches beside
//     it already existed to prevent.
//   - the eliminated-chooser fallback handed the gathered list to
//     applyFirstGatheredLocked unfiltered, so a question-asking effect
//     fired there too, unlike the mustSettleNow fallback next to it
//     which routes through skipQuestionsLocked.

const (
	optionalBonusOracle   = "00000000-0000-4000-8000-0000000d0090"
	optionalDoublerOracle = "00000000-0000-4000-8000-0000000d0091"
)

// optionalCounterBonus is "if counters would be put on a permanent,
// you MAY put ten more on it instead" — a CR 614.10 "may" with an
// arithmetic fingerprint, so a test can tell "the player said yes"
// from "the engine said yes for them".
//
// `fired` counts the Replace calls, which is what separates "skipped
// un-applied" from "applied and declined".
func optionalCounterBonus(label string, fired *int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			if fired != nil {
				*fired++
			}
			ev.CounterDelta += 10
			return nil
		},
		Optional:       true,
		Controller:     replacementSourceController,
		PromptQuestion: "Put ten more counters?",
		Label:          label,
	}
}

// expectOptionalPrompt asserts exactly one CR 614.10 yes/no prompt is
// queued, for `chooser`, and returns it.
func expectOptionalPrompt(t *testing.T, g *Game, chooser uuid.UUID) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want one CR 614.10 \"may\" prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", c.Kind, PendingChoiceOptionalReplacement)
	}
	if c.Chooser != chooser {
		t.Fatalf("chooser = %s, want %s", c.Chooser, chooser)
	}
	return c
}

// TestOptionalEffectInAChosenOrderIsStillOffered is the headline. Two
// effects in one window, so the affected player orders them; one of
// them is a "may", so its controller is still asked — and the answer
// decides, both ways round, with the rest of the chosen order applied
// after it in the order that was chosen.
//
// The arithmetic is the proof that the ORDER survived the pause:
// one counter with the bonus first is (1 + 10) × 2 = 22, and with the
// doubler first it is (1 × 2) + 10 = 12. Before the fix the prompt
// never appeared and both orders fired the bonus blind.
func TestOptionalEffectInAChosenOrderIsStillOffered(t *testing.T) {
	cases := []struct {
		name  string
		order []string
		apply bool
		want  int
	}{
		{"may first, accepted", []string{"May bonus", "Doubler"}, true, 22},
		{"may first, declined", []string{"May bonus", "Doubler"}, false, 2},
		{"may last, accepted", []string{"Doubler", "May bonus"}, true, 12},
		{"may last, declined", []string{"Doubler", "May bonus"}, false, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fired := 0
			stubCatalogReplacements(t, map[string][]ReplacementEffect{
				optionalBonusOracle:   {optionalCounterBonus("May bonus", &fired)},
				optionalDoublerOracle: {counterDoubler("Doubler")},
			})
			g := newActiveGame(t)
			owner := g.Seats[0].ID
			pushReplacementSource(g, optionalBonusOracle, owner)
			pushReplacementSource(g, optionalDoublerOracle, owner)
			bear := pushBear(g, owner)

			if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
				t.Fatalf("AddCounter: %v", err)
			}
			choice := replacementOrderPrompt(t, g)
			order := orderByLabel(t, g, choice, tc.order...)
			if err := g.ResolveReplacementOrder(choice.ID, choice.Chooser, order); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}

			// Nothing has landed yet: the chain paused on the "may".
			if got := countersOnCard(g, bear, "+1/+1"); got != 0 {
				t.Fatalf("counters = %d before the \"may\" is answered, want 0 — "+
					"the event lands from the resume, not from the ordering answer", got)
			}
			prompt := expectOptionalPrompt(t, g, owner)
			if err := g.ResolveOptionalReplacement(prompt.ID, prompt.Chooser, tc.apply); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}

			if got := countersOnCard(g, bear, "+1/+1"); got != tc.want {
				t.Errorf("counters = %d, want %d", got, tc.want)
			}
			wantFired := 0
			if tc.apply {
				wantFired = 1
			}
			if fired != wantFired {
				t.Errorf("the \"may\" fired %d times, want %d — a declined CR 614.10 "+
					"effect is marked applied without firing", fired, wantFired)
			}
			if len(g.PendingChoices) != 0 {
				t.Errorf("%d prompts left over after the chain settled", len(g.PendingChoices))
			}
		})
	}
}

// TestGoneChooserDropsAQuestionAskingEffect is the second gap. Nobody
// can answer the CR 616 ordering prompt, so the gathered order stands
// — but an effect that would ask its OWN question cannot ride that
// order either, because there is no order-free way to answer it. It is
// dropped un-applied (the weaker branch), and the rest of the window
// still applies.
//
// The "may" is gathered FIRST here, so the pre-#847 engine fired it:
// applyFirstGatheredLocked took the head of an unfiltered list.
func TestGoneChooserDropsAQuestionAskingEffect(t *testing.T) {
	fired := 0
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		optionalBonusOracle:   {optionalCounterBonus("May bonus", &fired)},
		optionalDoublerOracle: {counterDoubler("Doubler")},
	})
	g := newActiveGame(t)
	gone := g.Seats[1]
	pushReplacementSource(g, optionalBonusOracle, gone.ID)
	pushReplacementSource(g, optionalDoublerOracle, gone.ID)
	// The counters land on a permanent whose controller has left the
	// game, so there is nobody to ask for an order.
	bear := pushBear(g, gone.ID)
	gone.Eliminated = true

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("pending choices = %d, want 0 (the chooser is gone)", len(g.PendingChoices))
	}
	if fired != 0 {
		t.Errorf("the \"may\" fired %d times, want 0 — an effect that asks its own "+
			"question must not ride an order nobody chose", fired)
	}
	if got := countersOnCard(g, bear, "+1/+1"); got != 2 {
		t.Errorf("counters = %d, want 2 (the doubler alone)", got)
	}
}

// TestUndoAcrossTheOptionalPromptInAChosenOrderReplays is the undo
// contract for the new pause. Rewinding into the open "may" and
// answering it again has to land exactly what the first answer landed:
// the ordering answer's applied-effect marks and the in-flight event
// are both part of the snapshot (#793 / #808), so a replayed "yes"
// must not double the doubler it already applied.
func TestUndoAcrossTheOptionalPromptInAChosenOrderReplays(t *testing.T) {
	fired := 0
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		optionalBonusOracle:   {optionalCounterBonus("May bonus", &fired)},
		optionalDoublerOracle: {counterDoubler("Doubler")},
	})
	g := newActiveGame(t)
	owner := g.Seats[0].ID
	pushReplacementSource(g, optionalBonusOracle, owner)
	pushReplacementSource(g, optionalDoublerOracle, owner)
	bear := pushBear(g, owner)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	choice := replacementOrderPrompt(t, g)
	order := orderByLabel(t, g, choice, "Doubler", "May bonus")
	if err := g.ResolveReplacementOrder(choice.ID, choice.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	promptOpen := g.Clone()
	prompt := expectOptionalPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, prompt.Chooser, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	first := countersOnCard(g, bear, "+1/+1")
	if first != 12 {
		t.Fatalf("counters = %d on the first run, want 12 ((1 × 2) + 10)", first)
	}

	fired = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if got := countersOnCard(g, bear, "+1/+1"); got != 0 {
		t.Fatalf("counters = %d after rewinding into the open prompt, want 0", got)
	}
	replay := expectOptionalPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(replay.ID, replay.Chooser, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement (replay): %v", err)
	}
	if got := countersOnCard(g, bear, "+1/+1"); got != first {
		t.Errorf("replayed counters = %d, want %d — the same answer must land the same event", got, first)
	}
	if fired != 1 {
		t.Errorf("the \"may\" fired %d times on the replay, want 1", fired)
	}
}
