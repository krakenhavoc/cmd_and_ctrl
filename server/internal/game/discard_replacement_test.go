package game

import (
	"testing"

	"github.com/google/uuid"
)

// discard_replacement_test.go pins #650: a discard is its OWN
// replacement event, carrying who is discarding and why.
//
// #853 had already routed every discard through the shared exit, so
// the window opened — but what it opened was an anonymous hand →
// graveyard move. Library of Leng ("if an EFFECT causes you to
// discard"), madness and the Obstinate Baloth family all key on the
// discard and two of the three key on the cause, and none of them
// could be written against a move.
//
// The catalog half — the real Library of Leng — is in
// cards/effects/library_of_leng_test.go. What is here is the engine's
// own contract: the kind, the cause, the destination a replacement
// rewrites, the event that still fires afterwards, the CR 903.9 offer
// that still works through it, and the cost that still cannot pause.

// discardReplacement builds a test replacement on the discard event.
// `causes` empty means every cause.
func discardReplacement(label string, dst ZoneKind, optional bool, causes ...DiscardCause) ReplacementEffect {
	return ReplacementEffect{
		Watches:        []EventKind{EventDiscardCard},
		Optional:       optional,
		PromptQuestion: label,
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			if ev.Kind != RepEventDiscard || ev.NewZone == dst {
				return false
			}
			if len(causes) == 0 {
				return true
			}
			for _, c := range causes {
				if c == ev.DiscardCause {
					return true
				}
			}
			return false
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.NewZone = dst
			ev.NewZoneOwner = ev.DiscardPlayer
			return nil
		},
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID { return ev.DiscardPlayer },
		Label:      label,
	}
}

// handCard deals one ordinary card into p's hand and returns its ID.
func handCardForDiscard(t *testing.T, g *Game, p *Player, name string) uuid.UUID {
	t.Helper()
	c := NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// discardOne runs one effect-caused discard of `card` by p.
func discardOne(t *testing.T, g *Game, p *Player, card uuid.UUID, cause DiscardCause) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.discardCardsLocked(p.ID, []uuid.UUID{card}, discardOptions{cause: cause}); err != nil {
			t.Fatalf("discardCardsLocked: %v", err)
		}
	})
}

// --- the event's shape -------------------------------------------------

// TestADiscardOpensADiscardEventWithItsCause — the whole of #650 in one
// assertion: the window sees a discard, by whom, and why.
func TestADiscardOpensADiscardEventWithItsCause(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cause DiscardCause
	}{
		{"effect", DiscardCauseEffect},
		{"cleanup", DiscardCauseCleanup},
		{"cost", DiscardCauseCost},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[1]
			g.WithWriteLock(func() { p.Hand.Cards = nil })
			card := handCardForDiscard(t, g, p, "Pitch")

			var seen *ReplacementEvent
			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(ReplacementEffect{
					Watches: []EventKind{EventDiscardCard},
					AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
						cp := *ev
						seen = &cp
						return false
					},
					Label: "observer",
				})
			})
			discardOne(t, g, p, card, tc.cause)

			if seen == nil {
				t.Fatal("the discard opened no replacement window")
			}
			if seen.Kind != RepEventDiscard {
				t.Errorf("event kind = %q, want %q", seen.Kind, RepEventDiscard)
			}
			if seen.DiscardPlayer != p.ID {
				t.Errorf("DiscardPlayer = %s, want %s", seen.DiscardPlayer, p.ID)
			}
			if seen.DiscardCause != tc.cause {
				t.Errorf("DiscardCause = %q, want %q", seen.DiscardCause, tc.cause)
			}
			if seen.CardID != card || seen.OldZone != ZoneHand || seen.NewZone != ZoneGraveyard {
				t.Errorf("payload = %s %s → %s, want the card hand → graveyard", seen.CardID, seen.OldZone, seen.NewZone)
			}
		})
	}
}

// TestAnExileInsteadReplacementRedirectsADiscard — the mandatory,
// every-cause shape madness (#657) and the Obstinate Baloth family
// will use. The card goes to exile, and it was still DISCARDED
// (CR 701.8a defines the keyword action by the move out of the hand).
func TestAnExileInsteadReplacementRedirectsADiscard(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	card := handCardForDiscard(t, g, p, "Fiery Temper")
	w := watchDiscards(g)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(discardReplacement("exile it instead", ZoneExile, false))
	})

	discardOne(t, g, p, card, DiscardCauseEffect)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("a mandatory replacement prompted: %d choice(s)", len(g.PendingChoices))
	}
	assertOnlyIn(t, card, g.Exile, p.Hand, p.Graveyard)
	assertOneDiscardEvent(t, w, p.ID, card, ZoneExile)
	if w.discards[0].DiscardCause != DiscardCauseEffect {
		t.Errorf("EventDiscardCard cause = %q, want %q", w.discards[0].DiscardCause, DiscardCauseEffect)
	}
}

// TestACauseFilteredReplacementLeavesTheOtherCausesAlone — Library of
// Leng's clause in the abstract: Effect only, so a cost discard and a
// cleanup discard go to the graveyard untouched.
func TestACauseFilteredReplacementLeavesTheOtherCausesAlone(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cause   DiscardCause
		wantDst ZoneKind
	}{
		{"an effect's discard is replaced", DiscardCauseEffect, ZoneLibrary},
		{"a cost's discard is not", DiscardCauseCost, ZoneGraveyard},
		{"the cleanup discard is not", DiscardCauseCleanup, ZoneGraveyard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[1]
			g.WithWriteLock(func() { p.Hand.Cards = nil })
			card := handCardForDiscard(t, g, p, "Pitch")
			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(discardReplacement("to the library instead", ZoneLibrary, false, DiscardCauseEffect))
			})

			discardOne(t, g, p, card, tc.cause)

			switch tc.wantDst {
			case ZoneLibrary:
				assertOnlyIn(t, card, p.Library, p.Hand, p.Graveyard)
				if top := p.Library.Cards[len(p.Library.Cards)-1]; top.InstanceID != card {
					t.Error("the card did not land on TOP of the library")
				}
			default:
				assertOnlyIn(t, card, p.Graveyard, p.Hand, p.Library)
			}
		})
	}
}

// TestACostDiscardNeverPrompts — CR 601.2h pays a spell's costs as one
// indivisible step, so an OPTIONAL discard replacement is skipped
// un-applied rather than asked. Weaker than printed, never stronger,
// and never a wedged table.
func TestACostDiscardNeverPrompts(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	card := handCardForDiscard(t, g, p, "Pitch")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(discardReplacement("exile it instead?", ZoneExile, true))
	})

	discardOne(t, g, p, card, DiscardCauseCost)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("a cost discard queued %d prompt(s) — CR 601.2h pays costs as one step", len(g.PendingChoices))
	}
	assertOnlyIn(t, card, p.Graveyard, p.Hand, g.Exile)
}

// TestADiscardedCommanderStillGetsTheCR9039Offer — the built-in has to
// keep working through the new event kind, because "from anywhere"
// includes a hand and a discard is how a card leaves one.
func TestADiscardedCommanderStillGetsTheCR9039Offer(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	cmdID := emptyHandWithCommanders(t, g, p, 1)[0]
	w := watchDiscards(g)

	discardOne(t, g, p, cmdID, DiscardCauseEffect)

	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 903.9 prompt", len(g.PendingChoices))
	}
	if got := commanderPromptCard(t, g.PendingChoices[0]); got != cmdID {
		t.Errorf("the prompt is about %s, want the commander %s", got, cmdID)
	}
	answerOnlyCommanderPrompt(t, g, p.ID, true)
	assertOnlyIn(t, cmdID, p.Command, p.Hand, p.Graveyard)
	assertOneDiscardEvent(t, w, p.ID, cmdID, ZoneCommand)
}

// TestUndoAcrossAnOptionalDiscardReplacementReplaysTheSameWay — the
// "may" is answered an action later, so the undo stack has to rewind a
// game sitting on it and have the second answer land the same card in
// the same zone.
func TestUndoAcrossAnOptionalDiscardReplacementReplaysTheSameWay(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	card := handCardForDiscard(t, g, p, "Pitch")
	spare := handCardForDiscard(t, g, p, "Spare")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(discardReplacement("to the library instead?", ZoneLibrary, true, DiscardCauseEffect))
	})

	thenRuns := 0
	g.WithWriteLock(func() {
		if err := g.discardCardsLocked(p.ID, []uuid.UUID{card, spare}, discardOptions{
			cause: DiscardCauseEffect,
			then:  func(*Game) error { thenRuns++; return nil },
		}); err != nil {
			t.Fatalf("discardCardsLocked: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 614.10 prompt", len(g.PendingChoices))
	}

	promptOpen := g.Clone()
	answer := func() {
		t.Helper()
		for len(g.PendingChoices) > 0 {
			c := g.PendingChoices[0]
			if err := g.ResolveOptionalReplacement(c.ID, c.Chooser, true); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
		}
	}
	answer()
	if thenRuns != 1 {
		t.Fatalf(`first run: "then" ran %d times, want 1`, thenRuns)
	}
	assertOnlyIn(t, card, g.Seats[1].Library, g.Seats[1].Hand, g.Seats[1].Graveyard)

	thenRuns = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[1].Hand.Contains(card) || !g.Seats[1].Hand.Contains(spare) {
		t.Fatal("the rewind into the open prompt did not put both cards back in hand")
	}
	answer()
	if thenRuns != 1 {
		t.Errorf(`replay: "then" ran %d times, want 1`, thenRuns)
	}
	assertOnlyIn(t, card, g.Seats[1].Library, g.Seats[1].Hand, g.Seats[1].Graveyard)
}
