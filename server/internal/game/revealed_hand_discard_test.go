package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ADR 0116 (#2078): the revealed-hand pick's filter at the engine
// level. The card-level tests (cards/effects/revealed_hand_discard_test.go)
// run Thoughtseize, Unmask and Pelakka Predation through it with the
// real predicates; these pin the entry point's own rules.

// rhSetHand replaces p's hand with one land and n nonland cards and
// returns (land, nonlands).
func rhSetHand(p *Player, n int) (uuid.UUID, []uuid.UUID) {
	p.Hand.Cards = nil
	land := uuid.New()
	p.Hand.PushTop(Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: p.ID, Controller: p.ID})
	var spells []uuid.UUID
	for range n {
		id := uuid.New()
		p.Hand.PushTop(Card{InstanceID: id, Name: "Duress", TypeLine: "Sorcery", ManaCost: "{B}", Owner: p.ID, Controller: p.ID})
		spells = append(spells, id)
	}
	return land, spells
}

func rhNonland(c Card) bool { return !c.IsLand() }

func rhQueue(t *testing.T, g *Game, d RevealedHandDiscard) *PendingChoice {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.QueueDiscardFromRevealedHand(d) })
	if id == uuid.Nil {
		t.Fatal("no prompt was queued")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == id {
			return c
		}
	}
	t.Fatal("the queued prompt is not on the queue")
	return nil
}

// Count is capped at the number of cards the filter admits (CR 609.3):
// "choose two nonland cards" from a hand with one chooses one.
func TestRevealedHandDiscardCapsCountAtTheMatchingCards(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	_, spells := rhSetHand(victim, 1)

	c := rhQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 2, Reason: "Test",
		Filter: rhNonland, Label: "nonland card",
	})
	if c.Count != 1 || !slices.Equal(c.DiscardOptions, spells) {
		t.Errorf("count %d options %v, want 1 and %v", c.Count, c.DiscardOptions, spells)
	}
}

// ADR 0116 §5: a card named twice is refused, even when it is a legal
// card — a check the prompt never had.
func TestRevealedHandDiscardRefusesADuplicatePick(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	_, spells := rhSetHand(victim, 2)

	c := rhQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Count: 2, Reason: "Test",
		Filter: rhNonland, Label: "nonland card",
	})
	err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{spells[0], spells[0]})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("a duplicated pick: err = %v, want ErrInvalidParam", err)
	}
	if victim.Hand.Size() != 3 {
		t.Errorf("a refused pick moved cards: hand %d", victim.Hand.Size())
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, spells); err != nil {
		t.Fatalf("the two legal cards: %v", err)
	}
	if victim.Hand.Size() != 1 {
		t.Errorf("hand %d after discarding two of three", victim.Hand.Size())
	}
}

// A filter that admits nothing queues nothing, but the hand is still
// revealed to every seat (CR 609.3, 701.20a).
func TestRevealedHandDiscardWithNoMatchStillReveals(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	rhSetHand(victim, 0)
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueDiscardFromRevealedHand(RevealedHandDiscard{
			Chooser: chooser.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
			Filter: rhNonland, Label: "nonland card",
		})
	})
	if id != uuid.Nil || len(g.PendingChoices) != 0 {
		t.Fatalf("a prompt went up over a hand with no nonland card: %v", id)
	}
	for _, seat := range g.Seats {
		if !victim.Hand.Cards[0].IsKnownTo(seat.ID) {
			t.Errorf("%s did not see the revealed land", seat.Name)
		}
	}
}

// A prompt from before ADR 0116 (a restore point) carries no
// DiscardOptions, and still takes any card in the hand.
func TestALegacyRevealedHandPickTakesAnyCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	land, _ := rhSetHand(victim, 1)
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceDiscardFromHand, Chooser: chooser.ID, FromPlayer: victim.ID,
			Count: 1, Reason: "Thoughtseize",
		})
	})
	if err := g.ResolvePendingChoice(id, chooser.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("a legacy prompt refused a card in the hand: %v", err)
	}
}

// CR 800.4g, ADR 0060: when the chooser leaves and the prompt is
// reassigned, the new chooser inherits the same legal set — and, since
// the reveal went to the whole table, can read the hand.
//
// No printed card reaches this today: the revealed-hand pick is asked
// of the spell's controller, who takes the spell with them. The source
// here is a permanent another seat controls, which is the case the
// departure table's row exists for.
func TestAReassignedRevealedHandPickKeepsItsOptions(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim, controller := g.Seats[0], g.Seats[1], g.Seats[2]
	land, spells := rhSetHand(victim, 2)
	source := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: source, Name: "Source", TypeLine: "Enchantment", Owner: controller.ID, Controller: controller.ID})

	c := rhQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Source: source, Count: 1, Reason: "Test",
		Filter: rhNonland, Label: "nonland card",
	})
	if err := g.Concede(chooser.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	var moved *PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.ID == c.ID {
			moved = pc
		}
	}
	if moved == nil {
		t.Fatal("the prompt was dropped, not reassigned")
	}
	if moved.Chooser == chooser.ID || !slices.Equal(moved.DiscardOptions, spells) {
		t.Fatalf("reassigned to %v with options %v", moved.Chooser, moved.DiscardOptions)
	}
	for _, card := range victim.Hand.Cards {
		if !card.IsKnownTo(moved.Chooser) {
			t.Errorf("the new chooser cannot read %s", card.Name)
		}
	}
	if err := g.ResolvePendingChoice(moved.ID, moved.Chooser, []uuid.UUID{land}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("the new chooser took the land: %v", err)
	}
	if err := g.ResolvePendingChoice(moved.ID, moved.Chooser, spells[:1]); err != nil {
		t.Errorf("the new chooser's legal pick: %v", err)
	}
}
