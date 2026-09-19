package game

import (
	"testing"

	"github.com/google/uuid"
)

// affected_player_test.go — #982. CR 616.1 gives the ordering choice to
// "the affected object's controller or the affected player", never to
// whoever controls the replacements.
//
// `affectedPlayerForEvent` names that player per event kind and, when
// it has no case for a kind, falls through to the FIRST GATHERED
// EFFECT's controller — the effect whose source happens to sit earliest
// in battlefield order. Two of the kinds ADR 0061 added had no case:
// RepEventDiscard and RepEventCreateTokens.
//
// It was invisible because every catalog replacement of either kind is
// controller-scoped, so the fallback named the same player by accident.
// The shapes that break it are printed — Primal Vigor is deliberately
// symmetrical, and madness (#657) and the Obstinate Baloth family read
// a discard an OPPONENT caused — so the tests below build the board the
// fallback gets wrong: the first gathered effect belongs to somebody
// who is not the affected player.
//
// The structural half, which is what stops the next kind from being
// forgotten, is TestEveryReplacementEventKindIsSwitchedOn in
// replacement_kind_gate_test.go.

// foreignTokenDoubler is Primal Vigor's shape: every player's creation
// doubles, and the effect is controlled by `owner` — who is NOT the
// player creating the tokens in the tests below.
func foreignTokenDoubler(owner uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventTokenCreated},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCreateTokens && ev.TokenCount() > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.MultiplyTokens(2)
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

// foreignTokenKindSwapper is Academy Manufactor's shape with the same
// foreign controller: a second, DIFFERENT declared effect, so the
// window really is a CR 616.1 ordering question rather than #792's
// identical-window skip.
func foreignTokenKindSwapper(owner uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventTokenCreated},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventCreateTokens && ev.TokenCount() > 0
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.ReplaceTokenKinds(clueToken())
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

// foreignDiscardBecomes is the discard sibling: "if a player would
// discard a card, put it <dst> instead", controlled by `owner`.
func foreignDiscardBecomes(owner uuid.UUID, dst ZoneKind, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDiscardCard},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventDiscard && ev.NewZone != dst
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.NewZone = dst
			ev.NewZoneOwner = ev.DiscardPlayer
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

// onlyPrompt returns the single open prompt, failing if there is not
// exactly one.
func onlyPrompt(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open, want exactly 1", len(g.PendingChoices))
	}
	return g.PendingChoices[0]
}

// TestTheTokenCreatorOrdersTheCreationReplacements is Primal Vigor's
// bug: two replacements an OPPONENT controls apply to MY creation, so
// CR 616.1 asks ME which order they went in.
//
// Both injected effects are controlled by seat 1 and gathered before
// anything of seat 0's, so the pre-#982 fallback ("the first gathered
// effect's controller") answers seat 1 — the wrong player, and one who
// would then be holding a prompt about somebody else's tokens.
func TestTheTokenCreatorOrdersTheCreationReplacements(t *testing.T) {
	g := newActiveGame(t)
	creator, other := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(foreignTokenDoubler(other.ID, "their doubler"))
		g.RegisterReplacementForTest(foreignTokenKindSwapper(other.ID, "their swapper"))
	})

	createGoblins(t, g, creator.ID, 1, nil)

	p := onlyPrompt(t, g)
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
	}
	if p.Chooser != creator.ID {
		t.Errorf("chooser = %s, want the token controller %s (CR 616.1); got %s, "+
			"which is the controller of the replacements",
			p.Chooser, creator.ID, whichSeat(g, p.Chooser))
	}
}

// TestTheDiscardingPlayerOrdersTheDiscardReplacements is the other
// half, with the same board shape: CR 701.8a puts the card into the
// DISCARDING player's graveyard, so they are the affected player
// however the replacements are controlled.
func TestTheDiscardingPlayerOrdersTheDiscardReplacements(t *testing.T) {
	g := newActiveGame(t)
	discarder, other := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(foreignDiscardBecomes(other.ID, ZoneExile, "their exile instead"))
		g.RegisterReplacementForTest(foreignDiscardBecomes(other.ID, ZoneLibrary, "their library instead"))
	})
	card := handCardForDiscard(t, g, discarder, "Pitched Bear")

	discardOne(t, g, discarder, card, DiscardCauseEffect)

	p := onlyPrompt(t, g)
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
	}
	if p.Chooser != discarder.ID {
		t.Errorf("chooser = %s, want the discarding player %s (CR 701.8a); got %s, "+
			"which is the controller of the replacements",
			p.Chooser, discarder.ID, whichSeat(g, p.Chooser))
	}
}

// TestADroppedKeywordActionPromptStillRunsItsThen is the gap the
// enumeration test found: #976 gave a cancelled keyword action its
// terminal outcome (finishSettledReplacementLocked) and never gave one
// to an ABANDONED one, so a scry whose CR 616 ordering prompt was taken
// away dropped the rest of its sentence with the frame.
//
// "Scry 2, then draw a card" draws whether or not the scry happened —
// #808's rule for the life tail, #853's for the route, #762's for the
// tokens — and a prompt taken away is a terminal outcome too (§5j).
func TestADroppedKeywordActionPromptStillRunsItsThen(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scryer := g.Seats[1]
	g.WithWriteLock(func() {
		libraryCard(scryer, "Top Bear")
		libraryCard(scryer, "Next Bear")
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, func(n int) int { return n * 2 }, "twice"))
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionScry, func(n int) int { return n + 1 }, "one more"))
	})

	thenRuns := 0
	g.WithWriteLock(func() {
		if n := g.ScryThenForEffect(scryer.ID, uuid.Nil, 1, func(*Game) error {
			thenRuns++
			return nil
		}); n != 0 {
			t.Fatalf("ScryThenForEffect looked at %d cards, want 0 — the window should have paused", n)
		}
	})
	p := onlyPrompt(t, g)
	if p.Kind != PendingChoiceReplacementOrder || p.Chooser != scryer.ID {
		t.Fatalf("prompt = %q for %s, want a %q for the scrying player", p.Kind, p.Chooser, PendingChoiceReplacementOrder)
	}
	if thenRuns != 0 {
		t.Fatalf(`the "then" ran %d times while the prompt was still open`, thenRuns)
	}

	if err := g.Concede(scryer.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the concede", len(g.PendingChoices))
	}
	if thenRuns != 1 {
		t.Errorf(`the "then" ran %d times, want exactly 1 — an abandoned keyword action still owes `+
			`the rest of its sentence (#982)`, thenRuns)
	}
}

// whichSeat names a player for a failure message.
func whichSeat(g *Game, id uuid.UUID) string {
	for i, s := range g.Seats {
		if s.ID == id {
			return "seat " + string(rune('0'+i))
		}
	}
	return "nobody at the table"
}
