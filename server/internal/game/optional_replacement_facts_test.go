package game

import (
	"testing"

	"github.com/google/uuid"
)

// optional_replacement_facts_test.go — #2390. An optional_replacement
// prompt says what its yes does: a dredge's N, and where CR 903.9b's
// commander goes if its owner says no. Both are read off the paused
// event, so they are pinned against a real pause rather than a
// hand-built PendingChoice.

// optionalDrawReplacement is a "may" on a draw that declares a dredge
// count, standing in for the catalog's Dredge(n) without importing it.
func optionalDrawReplacement(n int) ReplacementEffect {
	return ReplacementEffect{
		Watches:  []EventKind{EventDrawCard},
		Optional: true,
		Dredge:   n,
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDraw && ev.DrawCount > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID {
			return ev.DrawPlayer
		},
		PromptQuestion: "Test draw replacement?",
		Label:          "test draw replacement",
	}
}

func TestDredgeOfferReadsTheEffectsCount(t *testing.T) {
	for _, n := range []int{0, 3} {
		g := newActiveGame(t)
		p := g.Seats[0]
		g.WithWriteLock(func() {
			g.testReplacements = append(g.testReplacements, optionalDrawReplacement(n))
		})
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
		prompt := onlyPrompt(t, g)
		if prompt.Kind != PendingChoiceOptionalReplacement {
			t.Fatalf("prompt kind %q, want %q", prompt.Kind, PendingChoiceOptionalReplacement)
		}
		if got := prompt.DredgeOffer(); got != n {
			t.Errorf("Dredge %d: DredgeOffer() = %d", n, got)
		}
		if got := prompt.CommanderHeadedFor(); got != "" {
			t.Errorf("a draw replacement is not the commander question, CommanderHeadedFor() = %q", got)
		}
	}
}

func TestCommanderHeadedForNamesTheDeclinedDestination(t *testing.T) {
	for _, tc := range []struct {
		name string
		move func(g *Game, id uuid.UUID) error
		want ZoneKind
	}{
		{"bounce", func(g *Game, id uuid.UUID) error { return g.BounceToHandForEffect(id) }, ZoneHand},
		{"tuck", func(g *Game, id uuid.UUID) error { return g.TuckToLibraryForEffect(id, false) }, ZoneLibrary},
	} {
		g := newActiveGame(t)
		owner := g.Seats[0]
		cmdID := seatCommander(t, g.Battlefield, owner)
		g.mu.Lock()
		err := tc.move(g, cmdID)
		g.mu.Unlock()
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		prompt := expectCommanderPrompt(t, g, owner)
		if got := prompt.CommanderHeadedFor(); got != tc.want {
			t.Errorf("%s: CommanderHeadedFor() = %q, want %q", tc.name, got, tc.want)
		}
		if got := prompt.DredgeOffer(); got != 0 {
			t.Errorf("%s: the commander question is not a dredge, DredgeOffer() = %d", tc.name, got)
		}
	}
}

// TestEntryPayLifePromptCarriesItsLife — the shockland's payment rides
// LifeCost beside its PayCost text, so legal can put #547's MoveCost on
// the pay branch.
func TestEntryPayLifePromptCarriesItsLife(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	land := Card{InstanceID: uuid.New(), Name: "Test Shock", TypeLine: "Land", Owner: p.ID, Controller: p.ID}
	g.WithWriteLock(func() {
		g.queueEntryPayLifePromptLocked(&ReplacementEvent{Kind: RepEventMove, CardID: land.InstanceID}, activeReplacement{
			effect: ReplacementEffect{EntryLifeCost: 2, Label: "Test Shock"},
		}, p.ID, 2)
	})
	prompt := onlyPrompt(t, g)
	if prompt.LifeCost != 2 || prompt.PayCost != "2 life" {
		t.Errorf("LifeCost %d, PayCost %q; want 2 and \"2 life\"", prompt.LifeCost, prompt.PayCost)
	}
}
