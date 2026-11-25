package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// revealed_hand_each_of_test.go — #6271: Distended Mindbender's "target
// opponent reveals their hand. You choose from it a nonland card with
// mana value 3 or less and a card with mana value 4 or greater. That
// player discards those cards." CR 701.20a (the reveal) and CR 701.9a
// (the discard).

func mindbenderFilters() []func(Card) bool {
	return []func(Card) bool{
		func(c Card) bool { return !c.IsLand() && c.ManaValue() <= 3 },
		func(c Card) bool { return c.ManaValue() >= 4 },
	}
}

func eachOfHandCard(p *Player, name, typeLine, cost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine, c.ManaCost = typeLine, cost
	p.Hand.PushTop(c)
	return c.InstanceID
}

// A 2-MV and a 5-MV card: both are chosen and discarded together; two
// cheap cards are not "a card for each".
func TestRevealedHandEachOfTakesOneCardPerFilter(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	cheap := eachOfHandCard(victim, "Cheap", "Sorcery", "{1}{B}")
	cheap2 := eachOfHandCard(victim, "Cheap Two", "Instant", "{U}")
	big := eachOfHandCard(victim, "Big", "Creature — Giant", "{3}{R}{R}")
	land := eachOfHandCard(victim, "Land", "Land", "")

	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Reason: "Test Mindbender",
		Label: "a nonland card with mana value 3 or less and a card with mana value 4 or greater", EachOf: mindbenderFilters(),
	})
	if c.Count != 2 {
		t.Fatalf("count = %d, want 2", c.Count)
	}
	for _, id := range c.DiscardOptions {
		if id == land {
			t.Fatal("a land is a candidate")
		}
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{cheap, cheap2}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("two cheap cards were taken: %v", err)
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{big, cheap}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	for _, id := range []uuid.UUID{cheap, big} {
		if !victim.Graveyard.Contains(id) || rhpEvents(g, EventDiscardCard, id) != 1 {
			t.Errorf("card %v was not discarded", id)
		}
	}
	if !victim.Hand.Contains(cheap2) || !victim.Hand.Contains(land) {
		t.Error("an unchosen card left the hand")
	}
}

// With only a 2-MV nonland card, only that one is chosen.
func TestRevealedHandEachOfWithOneFilterFilled(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	chooser, victim := g.Seats[0], g.Seats[1]
	cheap := eachOfHandCard(victim, "Cheap", "Sorcery", "{1}{B}")
	eachOfHandCard(victim, "Land", "Land", "")
	c := rhpQueue(t, g, RevealedHandDiscard{
		Chooser: chooser.ID, FromPlayer: victim.ID, Reason: "Test Mindbender", EachOf: mindbenderFilters(),
	})
	if c.Count != 1 {
		t.Fatalf("count = %d, want 1", c.Count)
	}
	if err := g.ResolvePendingChoice(c.ID, chooser.ID, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !victim.Graveyard.Contains(cheap) {
		t.Error("the cheap card was not discarded")
	}
}

// The filters are the pick's only restriction.
func TestRevealedHandEachOfRefusesAFilter(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var err error
	g.WithWriteLock(func() {
		_, err = g.RevealedHandPickForEffect(RevealedHandDiscard{
			Chooser: g.Seats[0].ID, FromPlayer: g.Seats[1].ID, Filter: rhNonland, EachOf: mindbenderFilters(),
		})
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("EachOf with a Filter: %v, want ErrInvalidParam", err)
	}
}

// The matching is a real one: a card that passes both filters stands
// for whichever the other card leaves.
func TestMaxSlotMatching(t *testing.T) {
	for _, tc := range []struct {
		slots []int
		n     int
		want  int
	}{
		{[]int{1, 1}, 2, 1},
		{[]int{3, 1}, 2, 2},
		{[]int{1, 2, 3}, 2, 2},
		{nil, 2, 0},
		{[]int{2}, 2, 1},
	} {
		if got := maxSlotMatching(tc.slots, tc.n); got != tc.want {
			t.Errorf("maxSlotMatching(%v, %d) = %d, want %d", tc.slots, tc.n, got, tc.want)
		}
	}
}
