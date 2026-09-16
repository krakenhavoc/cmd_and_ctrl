package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// choose_cards_validate_test.go — ChooseCardsPrompt.Validate (#624),
// the set-level rule a card-set pick could not express before: the
// resolver half. The enumerator half is legal/choose_cards_validate_test.go
// and the bot half is aiseat/choose_cards_validate_test.go.
//
// Every prompt here is test-only. The first card that needs the hook
// (Invasion of New Phyrexia's back face, #626) is not registered by
// this change.

// discardUnlessCreature is the rule the hook exists for — "discard two
// cards unless you discard a creature card" — as a test-only closure:
// two of anything, or exactly one creature.
func discardUnlessCreature(picked []Card) bool {
	switch len(picked) {
	case 2:
		return true
	case 1:
		return picked[0].IsCreature()
	}
	return false
}

// typedHand rewrites p's first three hand cards so the rule has
// something to bite on: a creature, then two non-creatures. Returns
// their IDs in that order.
func typedHand(t *testing.T, g *Game, p *Player) (creature, spellA, spellB uuid.UUID) {
	t.Helper()
	if p.Hand.Size() < 3 {
		t.Fatalf("setup: want 3 cards in hand, have %d", p.Hand.Size())
	}
	g.WithWriteLock(func() {
		p.Hand.Cards[0].TypeLine = "Creature — Bear"
		p.Hand.Cards[1].TypeLine = "Instant"
		p.Hand.Cards[2].TypeLine = "Sorcery"
	})
	return p.Hand.Cards[0].InstanceID, p.Hand.Cards[1].InstanceID, p.Hand.Cards[2].InstanceID
}

// TestChooseCardsValidateRejectsWithoutDequeuing — the core contract. A
// set the rule refuses is rejected with its own sentinel, the prompt
// stays open so the player can choose again, and Then does not run.
func TestChooseCardsValidateRejectsWithoutDequeuing(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	creature, spellA, spellB := typedHand(t, g, p)
	ran := 0
	var got []uuid.UUID
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  p.ID,
			Question: "Discard two cards unless you discard a creature card",
			Cards:    []uuid.UUID{creature, spellA, spellB},
			Min:      1,
			Max:      2,
			Zone:     ZoneHand,
			Validate: discardUnlessCreature,
			Then: func(_ *Game, picked []uuid.UUID) error {
				ran++
				got = picked
				return nil
			},
		})
	})
	id := g.PendingChoices[0].ID

	err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{spellA})
	if !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("one non-creature: error = %v, want ErrChoiceSetRejected", err)
	}
	// Distinct from the malformed-payload sentinel: the wire text is
	// the player's only explanation, and "invalid parameter" is not one.
	if errors.Is(err, ErrInvalidParam) {
		t.Error("a rule rejection must not read as ErrInvalidParam")
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].ID != id {
		t.Fatal("a rejected set dropped the prompt; the player can no longer choose again")
	}
	if ran != 0 {
		t.Fatalf("Then ran %d times for a rejected set", ran)
	}

	// The count bounds are still checked first, and still read as a
	// malformed answer rather than a rule.
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("empty pick under a floor of one: error = %v, want ErrInvalidParam", err)
	}

	// One creature is the "unless" branch.
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("a single creature was rejected: %v", err)
	}
	if ran != 1 || len(got) != 1 || got[0] != creature {
		t.Errorf("Then ran %d times with %v, want once with the creature", ran, got)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("an accepted set stayed in the queue")
	}
}

// TestChooseCardsValidateAcceptsAnyValidSet — two non-creatures pass
// the other branch of the rule, and Then receives them in submitted
// order.
func TestChooseCardsValidateAcceptsAnyValidSet(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	creature, spellA, spellB := typedHand(t, g, p)
	var got []uuid.UUID
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  p.ID,
			Cards:    []uuid.UUID{creature, spellA, spellB},
			Min:      1,
			Max:      2,
			Zone:     ZoneHand,
			Validate: discardUnlessCreature,
			Then: func(_ *Game, picked []uuid.UUID) error {
				got = picked
				return nil
			},
		})
	})
	id := g.PendingChoices[0].ID
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{spellB, spellA}); err != nil {
		t.Fatalf("two cards rejected: %v", err)
	}
	if len(got) != 2 || got[0] != spellB || got[1] != spellA {
		t.Errorf("Then got %v, want [spellB spellA] in submitted order", got)
	}
}

// TestChooseCardsValidateIsSkippedForTheEmptyPick — a zero-floor
// prompt's "choose nothing" is the answer the enumerator marks
// AlwaysLegal, so no rule may refuse it. A hook that refuses
// everything proves it is not even consulted.
func TestChooseCardsValidateIsSkippedForTheEmptyPick(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	hand := handIDs(p, 2)
	calls := 0
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser: p.ID,
			Cards:   hand,
			Min:     0,
			Zone:    ZoneHand,
			Validate: func([]Card) bool {
				calls++
				return false
			},
			Then: func(*Game, []uuid.UUID) error { return nil },
		})
	})
	id := g.PendingChoices[0].ID
	if err := g.ResolveChooseCards(id, p.ID, hand[:1]); !errors.Is(err, ErrChoiceSetRejected) {
		t.Fatalf("non-empty pick: error = %v, want ErrChoiceSetRejected", err)
	}
	callsBefore := calls
	if err := g.ResolveChooseCards(id, p.ID, nil); err != nil {
		t.Fatalf("choose nothing was refused: %v", err)
	}
	if calls != callsBefore {
		t.Error("Validate was consulted for the empty pick")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("choose nothing did not clear the prompt")
	}
}

// TestChooseCardsValidateSeesTheLiveCards — the hook reads the cards as
// they are when the answer arrives, not as they were when the prompt
// was queued, and it runs after the zone re-check, so a card that has
// left is ErrCardNotFound before the rule is ever asked.
func TestChooseCardsValidateSeesTheLiveCards(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	creature, spellA, spellB := typedHand(t, g, p)
	var seen []string
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser: p.ID,
			Cards:   []uuid.UUID{creature, spellA, spellB},
			Min:     1,
			Max:     2,
			Zone:    ZoneHand,
			Validate: func(picked []Card) bool {
				seen = seen[:0]
				for _, c := range picked {
					seen = append(seen, c.TypeLine)
				}
				return discardUnlessCreature(picked)
			},
			Then: func(*Game, []uuid.UUID) error { return nil },
		})
		// The board moves while the prompt is open: the instant is
		// now (for the purposes of this test) a creature, and the
		// sorcery is discarded.
		p.Hand.Cards[1].TypeLine = "Artifact Creature — Golem"
		if _, err := MoveCard(p.Hand, p.Graveyard, spellB); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	id := g.PendingChoices[0].ID

	seen = nil
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{spellB}); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("a card that left the hand: error = %v, want ErrCardNotFound", err)
	}
	if len(seen) != 0 {
		t.Error("Validate ran for a pick that failed the zone re-check")
	}
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{spellA}); err != nil {
		t.Fatalf("the card that became a creature was rejected: %v", err)
	}
	if len(seen) != 1 || seen[0] != "Artifact Creature — Golem" {
		t.Errorf("Validate saw %v, want the live type line", seen)
	}
}

// TestChooseCardsPickLegalLockedAgreesWithTheResolver — the enumerator's
// window must say exactly what the resolver does, on the live game and
// on a clone (legal's dispatchAll answers against clones, and the
// clone shares the continuation frame and so the hook).
func TestChooseCardsPickLegalLockedAgreesWithTheResolver(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	creature, spellA, spellB := typedHand(t, g, p)
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  p.ID,
			Cards:    []uuid.UUID{creature, spellA, spellB},
			Min:      1,
			Max:      2,
			Zone:     ZoneHand,
			Validate: discardUnlessCreature,
			Then:     func(*Game, []uuid.UUID) error { return nil },
		})
	})
	sets := [][]uuid.UUID{
		{creature},
		{spellA},
		{spellB},
		{creature, spellA},
		{spellA, spellB},
		{spellA, spellA},
		{},
		{creature, spellA, spellB},
	}
	for _, set := range sets {
		var legal bool
		g.ReadSnapshot(func() {
			legal = g.ChooseCardsPickLegalLocked(g.PendingChoices[0], set)
		})
		clone := g.Clone()
		err := clone.ResolveChooseCards(clone.PendingChoices[0].ID, p.ID, set)
		if legal != (err == nil) {
			t.Errorf("set of %d: ChooseCardsPickLegalLocked=%v but ResolveChooseCards on a clone = %v",
				len(set), legal, err)
		}
	}
	// A search prompt is not this window's business.
	var other bool
	g.ReadSnapshot(func() {
		other = g.ChooseCardsPickLegalLocked(&PendingChoice{Kind: PendingChoiceSearchLibrary}, nil)
	})
	if other {
		t.Error("ChooseCardsPickLegalLocked accepted a pick for a different kind of prompt")
	}
}
