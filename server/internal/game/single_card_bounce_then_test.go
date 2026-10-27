package game

import (
	"testing"

	"github.com/google/uuid"
)

// single_card_bounce_then_test.go — #993's engine half.
//
// The single-card bounce with a continuation is a batch of one through
// BounceCardsToHandKeyedForEffect since #4253's cleanup deleted its closure
// wrapper, BounceToHandThenForEffect (the fourth beside
// ExileCardThenForEffect, SacrificeThenForEffect and
// TuckToLibraryThenForEffect). It exists because a
// hand is a CR 903.9b destination, so EVERY bounce can pause, and the
// catalog had no way to say "return it, THEN …" for one card.
//
// These are exiled_this_way_test.go's single-card cases with the
// destination changed, and they assert the same three things: the
// continuation does not run while the prompt is open, it runs exactly
// once when the answer comes in, and what it is told is CR 400.7's
// reading — the object that ARRIVED in a hand.

// TestSingleCardBounceThenReportsWhatLanded. A commander returned to
// its owner's hand stops to ask them about the command zone; a
// commander that takes the offer left the battlefield but did not
// reach a hand, so it was not returned this way.
func TestSingleCardBounceThenReportsWhatLanded(t *testing.T) {
	for _, tc := range []struct {
		name        string
		commandZone bool
		want        bool
	}{
		{"to the command zone", true, false},
		{"to the hand", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			owner := g.Seats[0]
			commander := seatCommander(t, g.Battlefield, owner)

			resetProbes()
			g.WithWriteLock(func() {
				err := g.BounceCardsToHandKeyedForEffect([]uuid.UUID{commander}, probeTail())
				if err != nil {
					t.Fatalf("BounceCardsToHandKeyedForEffect: %v", err)
				}
			})
			if ran := len(testProbeHeard); ran != 0 {
				t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0 — "+
					"a clause written on the next line is the #993 bug", ran)
			}

			prompt := expectCommanderPrompt(t, g, owner)
			if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, tc.commandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			heard := probeExiled()
			if ran := len(heard); ran != 1 {
				t.Fatalf("the continuation ran %d times after the answer, want 1", ran)
			}
			if got := heard[0]; got != tc.want {
				t.Errorf("bounced = %v, want %v — CR 400.7, the object that arrived", got, tc.want)
			}
		})
	}
}

// TestSingleCardBounceThenIsToldAboutACancelledMove — the other way a
// bounce fails to land, and the one with no prompt in it. The
// continuation still runs: a caller that is waiting has to be told even
// when the answer is "nothing happened".
func TestSingleCardBounceThenIsToldAboutACancelledMove(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	saved := wipeTarget(g, owner)

	resetProbes()
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(saved,
			"it can't be returned to a hand", func(ev *ReplacementEvent) { ev.Cancel() }))
		err := g.BounceCardsToHandKeyedForEffect([]uuid.UUID{saved}, probeTail())
		if err != nil {
			t.Fatalf("BounceCardsToHandKeyedForEffect: %v", err)
		}
	})
	heard := probeExiled()
	ran, got := len(heard), len(heard) > 0 && heard[0]
	if ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a cancelled move is still an answer", ran)
	}
	if got {
		t.Error("bounced = true for a move the CR 614 window cancelled, want false")
	}
	if !g.Battlefield.Contains(saved) {
		t.Error("the cancelled move leaves the permanent on the battlefield")
	}
}
