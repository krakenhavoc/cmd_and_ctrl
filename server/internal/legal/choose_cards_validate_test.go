package legal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choose_cards_validate_test.go — the enumerator half of #624. A
// choose-cards prompt may carry a set-level Validate hook on its
// unexported continuation frame, which is exactly the shape that
// wedged a bot seat in #544 when search grew one: the enumerator could
// not see the rule, offered what the resolver refuses, and a seat owing
// a choice is offered nothing else. So the promise, pinned here, is the
// search one: every offered set passes the engine's own check, and the
// budget REACHES the valid sets rather than being spent on the invalid
// ones.
//
// The prompts are test-only; no registered card uses the hook yet
// (Invasion of New Phyrexia, #626, is the first that will).

// unlessCreature is "discard two cards unless you discard a creature
// card": two of anything, or exactly one creature.
func unlessCreature(picked []game.Card) bool {
	switch len(picked) {
	case 2:
		return true
	case 1:
		return picked[0].IsCreature()
	}
	return false
}

// handOfTypes replaces the seat's hand with one card per type line, in
// order, and returns their IDs.
func handOfTypes(p *game.Player, typeLines ...string) []uuid.UUID {
	clearHand(p)
	ids := make([]uuid.UUID, 0, len(typeLines))
	for i, tl := range typeLines {
		// handCard pushes on top; the prompt's candidate order is
		// ids, which is what the enumerator walks.
		ids = append(ids, handCard(p, game.Card{Name: fmt.Sprintf("Card%d", i), TypeLine: tl}))
	}
	return ids
}

func queueDiscardUnlessCreature(g *game.Game, seat uuid.UUID, cards []uuid.UUID) {
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  seat,
			Question: "Discard two cards unless you discard a creature card",
			Cards:    cards,
			Min:      1,
			Max:      2,
			Zone:     game.ZoneHand,
			Validate: unlessCreature,
			Then:     func(*game.Game, []uuid.UUID) error { return nil },
		})
	})
}

// TestChooseCardsNeverOffersASetValidateRejects is the invariant:
// every enumerated answer is one the engine accepts. dispatchAll is
// the assertion; the checks after it say out loud which rejection this
// is about, so a regression reads as "offered a lone Instant" rather
// than an opaque error.
func TestChooseCardsNeverOffersASetValidateRejects(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	cards := handOfTypes(me, "Instant", "Creature — Bear", "Sorcery", "Land", "Creature — Elf")
	queueDiscardUnlessCreature(g, me.ID, cards)
	choice := g.PendingChoices[0]

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	types := map[uuid.UUID]string{}
	for _, c := range me.Hand.Cards {
		types[c.InstanceID] = c.TypeLine
	}
	singles, pairs := 0, 0
	for _, m := range moves {
		if m.AlwaysLegal {
			t.Errorf("a prompt with a floor of one marked %q always-legal", m.Label)
		}
		ids := pickedCardIDs(t, m)
		var ok bool
		g.ReadSnapshot(func() { ok = g.ChooseCardsPickLegalLocked(choice, ids) })
		if !ok {
			t.Errorf("offered %q, which ChooseCardsPickLegalLocked refuses", m.Label)
		}
		switch len(ids) {
		case 1:
			singles++
			if types[ids[0]] != "Creature — Bear" && types[ids[0]] != "Creature — Elf" {
				t.Errorf("offered a lone non-creature: %q", m.Label)
			}
		case 2:
			pairs++
		}
	}
	if singles != 2 {
		t.Errorf("offered %d single-card answers, want the 2 creatures: %v", singles, labels(moves))
	}
	if pairs == 0 {
		t.Errorf("no two-card answer was offered: %v", labels(moves))
	}
}

// TestChooseCardsReachesValidPairsPastTheBudget mirrors
// TestSearchForUpToTwoReachesValidPairsAtEveryLibrarySize, and it is
// the half a filter alone does not fix.
//
// combinations() fills MaxExpansionPerSource (12) smallest size first.
// For a "one creature, or two cards" prompt over a hand with no
// creature, every single is invalid: at 12 or more candidates the
// budget is all singles, filtering them leaves an EMPTY list, and a
// seat owing a choice with an empty list is the #544 wedge. So the
// sizes bracket the cap, both with no creature at all and with the
// only creature placed last, where a smallest-first walk would meet it
// only after the other singles.
func TestChooseCardsReachesValidPairsPastTheBudget(t *testing.T) {
	for _, n := range []int{3, 11, 12, 13, 20, 40} {
		for _, lastIsCreature := range []bool{false, true} {
			name := fmt.Sprintf("%d_cards_no_creature", n)
			if lastIsCreature {
				name = fmt.Sprintf("%d_cards_creature_last", n)
			}
			t.Run(name, func(t *testing.T) {
				g := newTable(t)
				me := g.Seats[0]
				typeLines := make([]string, n)
				for i := range typeLines {
					typeLines[i] = "Instant"
				}
				if lastIsCreature {
					typeLines[n-1] = "Creature — Bear"
				}
				cards := handOfTypes(me, typeLines...)
				queueDiscardUnlessCreature(g, me.ID, cards)

				moves := legal.EnumerateFor(g, me.ID)
				if len(moves) == 0 {
					t.Fatalf("a seat owing a satisfiable prompt was offered NOTHING — the #544 wedge")
				}
				dispatchAll(t, g, me.ID, moves)

				singles, pairs := 0, 0
				for _, m := range moves {
					switch len(pickedCardIDs(t, m)) {
					case 1:
						singles++
					case 2:
						pairs++
					}
				}
				t.Logf("candidates=%d answers=%d singles=%d pairs=%d", n, len(moves), singles, pairs)
				if pairs == 0 {
					t.Errorf("no valid two-card answer was reached: %d singles, 0 pairs", singles)
				}
				if lastIsCreature && singles != 1 {
					t.Errorf("the lone creature is a legal answer on its own and must be on offer; got %d singles", singles)
				}
				if !lastIsCreature && singles != 0 {
					t.Errorf("offered %d singles from a hand with no creature", singles)
				}
			})
		}
	}
}

// TestChooseCardsWithValidateKeepsChooseNothingAlwaysLegal — the
// engine skips Validate for the empty pick, so a zero-floor prompt's
// "choose nothing" stays the one answer marked AlwaysLegal even under a
// rule that refuses every non-empty set. The seat is never left with
// nothing.
func TestChooseCardsWithValidateKeepsChooseNothingAlwaysLegal(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	cards := handOfTypes(me, "Instant", "Sorcery", "Land")
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "Reveal any number",
			Cards:    cards,
			Min:      0,
			Zone:     game.ZoneHand,
			Validate: func([]game.Card) bool { return false },
			Then:     func(*game.Game, []uuid.UUID) error { return nil },
		})
	})
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 1 {
		t.Fatalf("enumerated %d answers, want only choose nothing: %v", len(moves), labels(moves))
	}
	if !moves[0].AlwaysLegal || len(pickedCardIDs(t, moves[0])) != 0 {
		t.Errorf("the one answer is %q (always_legal=%v), want the empty pick marked always-legal",
			moves[0].Label, moves[0].AlwaysLegal)
	}
	dispatchAll(t, g, me.ID, moves)
}

// TestChooseCardsSynthesisesNoFallback — a floor-one prompt whose rule
// accepts nothing gets no invented answer: anything offered would be
// refused, which is the #543 shape, and marking one AlwaysLegal would
// be a lie the runner's rescue then acts on. The enumerator logs it
// instead (a queuing effect must not ask an unanswerable question; see
// ChooseCardsPrompt.Validate).
func TestChooseCardsSynthesisesNoFallback(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	cards := handOfTypes(me, "Instant", "Sorcery")
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "Discard a creature card",
			Cards:    cards,
			Min:      1,
			Max:      1,
			Zone:     game.ZoneHand,
			Validate: unlessCreature,
			Then:     func(*game.Game, []uuid.UUID) error { return nil },
		})
	})
	if moves := legal.EnumerateFor(g, me.ID); len(moves) != 0 {
		t.Errorf("offered %v for a prompt no set satisfies", labels(moves))
	}
}
