package game

import (
	"testing"

	"github.com/google/uuid"
)

// replacement_recheck_test.go pins #802 (fixed on develop by #808's CR 616.1f loop): CR 616.1 applies ONE
// replacement and then looks again at the modified event, so an
// effect whose condition stopped holding partway through a chain must
// not fire. Both order paths re-check (stillAppliesLocked; the engine
// path fires one and re-gathers, CR 616.1f), so both are tested here: the order
// the affected player submits and the gathered order the engine falls
// back on.
//
// The canonical shape is a threshold clause behind a reducer — "if
// three or more counters would be put on a permanent, put twice that
// many instead" ordered after "put two fewer instead". Three counters
// reduced to one is no longer three or more, and the doubler has
// nothing to double.

const (
	recheckThresholdOracle = "00000000-0000-4000-8000-0000000d0080"
	recheckReducerOracle   = "00000000-0000-4000-8000-0000000d0081"
	recheckAmbientOracle   = "00000000-0000-4000-8000-0000000d0082"
	recheckLoyalOracle     = "00000000-0000-4000-8000-0000000d0083"
)

// counterThresholdDoubler is "if three or more counters would be put
// on a permanent, put twice that many on it instead". Its AppliesTo
// reads the event's own delta, which is what an earlier effect in the
// chain can take away from it.
func counterThresholdDoubler(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterDelta >= 3
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Controller: replacementSourceController,
		Label:      label,
	}
}

// counterReducerTwo is "if counters would be put on a permanent, put
// two fewer on it instead".
func counterReducerTwo(label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCounter && ev.CounterDelta > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta -= 2
			return nil
		},
		Controller: replacementSourceController,
		Label:      label,
	}
}

// replacementSourceController is the Controller hook every
// catalog-shaped replacement here shares: a static ability is
// controlled by whoever controls its source.
func replacementSourceController(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID {
	if src == nil {
		return uuid.Nil
	}
	return src.Controller
}

// replacementOrderPrompt returns the single queued CR 616 ordering
// prompt, failing the test when there isn't exactly one.
func replacementOrderPrompt(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 CR 616 ordering prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("pending choice kind = %q, want %q", c.Kind, PendingChoiceReplacementOrder)
	}
	return c
}

// orderByLabel builds the ID order a client would submit, naming the
// effects by the labels the prompt renders. Uses the same label
// lookup the wire projection does, so a mis-minted ID shows up here.
func orderByLabel(t *testing.T, g *Game, choice *PendingChoice, labels ...string) []ReplacementEffectID {
	t.Helper()
	byLabel := make(map[string]ReplacementEffectID, len(choice.ReplacementEffectIDs))
	g.mu.Lock()
	for _, id := range choice.ReplacementEffectIDs {
		label, _ := g.ReplacementOptionMetaForEffect(id)
		byLabel[label] = id
	}
	g.mu.Unlock()
	out := make([]ReplacementEffectID, 0, len(labels))
	for _, want := range labels {
		id, ok := byLabel[want]
		if !ok {
			t.Fatalf("prompt has no option labelled %q (has %v)", want, byLabel)
		}
		out = append(out, id)
	}
	return out
}

// TestChosenOrderReChecksAppliesToBetweenApplications is the headline
// for the player-chosen order, both ways round. The affected player
// orders a threshold doubler and a reducer; putting the reducer first
// breaks the threshold and the doubler drops out, which is CR 616.1
// looking at the modified event rather than the one it gathered.
//
// It is also the paused-window case: the window pauses on the
// ordering prompt and the re-check happens when the answer resumes
// it, on the effects that are still to run.
func TestChosenOrderReChecksAppliesToBetweenApplications(t *testing.T) {
	cases := []struct {
		name  string
		order []string
		want  int
	}{
		{
			// 3 → 6 (doubled) → 4 (two fewer).
			name:  "threshold first, both apply",
			order: []string{"Threshold doubler", "Reducer"},
			want:  4,
		},
		{
			// 3 → 1 (two fewer) → the threshold is no longer met.
			name:  "reducer first, threshold drops out",
			order: []string{"Reducer", "Threshold doubler"},
			want:  1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubCatalogReplacements(t, map[string][]ReplacementEffect{
				recheckThresholdOracle: {counterThresholdDoubler("Threshold doubler")},
				recheckReducerOracle:   {counterReducerTwo("Reducer")},
			})
			g := newActiveGame(t)
			owner := g.Seats[0].ID
			pushReplacementSource(g, recheckThresholdOracle, owner)
			pushReplacementSource(g, recheckReducerOracle, owner)
			bear := pushBear(g, owner)

			if err := g.AddCounter(bear, "+1/+1", 3); err != nil {
				t.Fatalf("AddCounter: %v", err)
			}
			choice := replacementOrderPrompt(t, g)
			order := orderByLabel(t, g, choice, tc.order...)

			if err := g.ResolveReplacementOrder(choice.ID, choice.Chooser, order); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			if got := countersOnCard(g, bear, "+1/+1"); got != tc.want {
				t.Errorf("counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestGatheredOrderReChecksAppliesToBetweenApplications is the same
// question asked of the engine-order path — the one taken when the
// CR 616 prompt is skipped because nobody can answer it. The gathered
// order is battlefield order, so pushing the sources the other way
// round is how the test picks the order.
func TestGatheredOrderReChecksAppliesToBetweenApplications(t *testing.T) {
	cases := []struct {
		name  string
		first string
		want  int
	}{
		{"threshold first, both apply", recheckThresholdOracle, 4},
		{"reducer first, threshold drops out", recheckReducerOracle, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubCatalogReplacements(t, map[string][]ReplacementEffect{
				recheckThresholdOracle: {counterThresholdDoubler("Threshold doubler")},
				recheckReducerOracle:   {counterReducerTwo("Reducer")},
			})
			g := newActiveGame(t)
			owner, gone := g.Seats[0].ID, g.Seats[1]
			second := recheckReducerOracle
			if tc.first == recheckReducerOracle {
				second = recheckThresholdOracle
			}
			pushReplacementSource(g, tc.first, owner)
			pushReplacementSource(g, second, owner)
			// The counters land on a permanent whose controller has
			// left the game, so there is nobody to ask for an order
			// and the engine applies the gathered one.
			bear := pushBear(g, gone.ID)
			gone.Eliminated = true

			if err := g.AddCounter(bear, "+1/+1", 3); err != nil {
				t.Fatalf("AddCounter: %v", err)
			}
			if len(g.PendingChoices) != 0 {
				t.Fatalf("pending choices = %d, want 0 (the chooser is gone)", len(g.PendingChoices))
			}
			if got := countersOnCard(g, bear, "+1/+1"); got != tc.want {
				t.Errorf("counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestResumeReChecksAgainstTheBoardAsItIsNow is the other half of the
// re-check: an effect can stop applying because the GAME moved under
// it while the prompt was outstanding, not only because an earlier
// effect in the chain rewrote the event. A replacement that only
// touches counters on a permanent its own controller controls stops
// applying the moment that permanent changes hands.
func TestResumeReChecksAgainstTheBoardAsItIsNow(t *testing.T) {
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		recheckAmbientOracle: {{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: replacementSourceController,
			Label:      "Ambient doubler",
		}},
		recheckLoyalOracle: {counterDoubler("Loyal doubler")},
	})
	g := newActiveGame(t)
	mine, theirs := g.Seats[0].ID, g.Seats[1].ID
	pushReplacementSource(g, recheckAmbientOracle, mine)
	pushReplacementSource(g, recheckLoyalOracle, mine)
	bear := pushBear(g, mine)

	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	choice := replacementOrderPrompt(t, g)
	order := orderByLabel(t, g, choice, "Ambient doubler", "Loyal doubler")

	// The bear changes hands while the prompt sits unanswered.
	g.mu.Lock()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].Controller = theirs
		}
	}
	g.mu.Unlock()

	if err := g.ResolveReplacementOrder(choice.ID, choice.Chooser, order); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := countersOnCard(g, bear, "+1/+1"); got != 2 {
		t.Errorf("counters = %d, want 2 (only the ambient doubler still applies)", got)
	}
}
