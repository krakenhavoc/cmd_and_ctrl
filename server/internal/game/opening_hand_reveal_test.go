package game

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

// opening_hand_reveal_test.go pins the reveal shape of CR 103.6 (#6101):
// a reveal arms a trigger at its owner's first upkeep, the starting
// player's included, and a decline arms nothing.

const revealTestOracle = "test-opening-hand-reveal"

// revealTable starts a four-seat game with the mulligan window open and the
// catalog answering revealTestOracle with a reveal whose body counts its
// resolutions by controller.
func revealTable(t *testing.T) (*Game, map[uuid.UUID]int) {
	t.Helper()
	fired := map[uuid.UUID]int{}
	body := testBody(func(_ *Game, item *StackItem) error {
		fired[item.Controller]++
		return nil
	})
	prev := CatalogOpeningHand
	CatalogOpeningHand = func(key string) *OpeningHandAction {
		if key == revealTestOracle {
			return &OpeningHandAction{Reveal: &OpeningHandReveal{Label: "Test — first upkeep", Body: body}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { CatalogOpeningHand = prev })
	g := NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]Card, 20)
		for j := range deck {
			deck[j] = NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, fired
}

// revealCardInHand puts the reveal card into a seat's hand.
func revealCardInHand(g *Game, seat int) uuid.UUID {
	p := g.Seats[seat]
	c := Card{InstanceID: uuid.New(), Name: "Revealer", TypeLine: "Creature — Eldrazi", OracleID: revealTestOracle, Owner: p.ID, Controller: p.ID}
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

func revealKeepAll(t *testing.T, g *Game) {
	t.Helper()
	for i := range g.Seats {
		if err := g.KeepHand(g.Seats[(g.StartingSeat+i)%len(g.Seats)].ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
}

func revealOffer(t *testing.T, g *Game, seat uuid.UUID) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind == PendingChoiceConfirm && c.Chooser == seat {
			return c
		}
	}
	t.Fatalf("no reveal offer for %v: %+v", seat, g.PendingChoices)
	return nil
}

// resolveStack passes priority until the stack is empty.
func revealResolveStack(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 64 && (len(g.StackMeta) > 0 || len(g.PendingTriggers) > 0); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// The starting player's first upkeep has already begun when the answer
// arrives (ADR 0133 decision 3), so the reveal's trigger goes on the stack
// in that upkeep; the card stays in hand, revealed to every seat.
func TestOpeningHandRevealTriggersAtTheStartingPlayersFirstUpkeep(t *testing.T) {
	g, fired := revealTable(t)
	starter := g.Seats[g.StartingSeat]
	card := revealCardInHand(g, g.StartingSeat)
	revealKeepAll(t, g)
	offer := revealOffer(t, g, starter.ID)
	if !offer.private {
		t.Error("the reveal question names a card in a hidden zone, so it is private")
	}
	if err := g.ResolveConfirm(offer.ID, starter.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	if !starter.Hand.Contains(card) {
		t.Fatal("revealing moved the card out of the hand")
	}
	for _, p := range g.Seats {
		if c, _ := g.cardInZoneLocked(starter.Hand, card); !c.KnownBy[p.ID] {
			t.Errorf("seat %s does not know the revealed card", p.Name)
		}
	}
	if g.Turn.Seq != 1 || g.Turn.Step != StepUpkeep {
		t.Fatalf("turn %d step %s, want turn 1's upkeep", g.Turn.Seq, g.Turn.Step)
	}
	revealResolveStack(t, g)
	if fired[starter.ID] != 1 {
		t.Fatalf("the trigger resolved %d times for the starting player, want 1", fired[starter.ID])
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a trigger is still waiting: %+v", g.DelayedTriggers)
	}
}

// Another seat's first upkeep is still to come: the trigger waits for it
// and fires there, once, and not in anyone else's upkeep.
func TestOpeningHandRevealWaitsForAnotherSeatsFirstUpkeep(t *testing.T) {
	g, fired := revealTable(t)
	seatIdx := (g.StartingSeat + 1) % len(g.Seats)
	seat := g.Seats[seatIdx]
	revealCardInHand(g, seatIdx)
	revealKeepAll(t, g)
	if err := g.ResolveConfirm(revealOffer(t, g, seat.ID).ID, seat.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	revealResolveStack(t, g)
	if fired[seat.ID] != 0 {
		t.Fatal("the trigger fired in the starting player's upkeep")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 waiting", len(g.DelayedTriggers))
	}
	for i := 0; i < 40 && !(g.Turn.ActiveSeat == seatIdx && g.Turn.Step == StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != seatIdx || g.Turn.Step != StepUpkeep {
		t.Fatalf("never reached the seat's upkeep: seat %d step %s", g.Turn.ActiveSeat, g.Turn.Step)
	}
	revealResolveStack(t, g)
	if fired[seat.ID] != 1 {
		t.Fatalf("the trigger resolved %d times in the seat's first upkeep, want 1", fired[seat.ID])
	}
}

// "You may": a decline reveals nothing and arms nothing.
func TestOpeningHandRevealDeclinedArmsNothing(t *testing.T) {
	g, fired := revealTable(t)
	starter := g.Seats[g.StartingSeat]
	card := revealCardInHand(g, g.StartingSeat)
	revealKeepAll(t, g)
	if err := g.ResolveConfirm(revealOffer(t, g, starter.ID).ID, starter.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	revealResolveStack(t, g)
	if fired[starter.ID] != 0 || len(g.DelayedTriggers) != 0 {
		t.Errorf("a decline armed the trigger: fired %d, waiting %d", fired[starter.ID], len(g.DelayedTriggers))
	}
	for _, p := range g.Seats {
		if c, _ := g.cardInZoneLocked(starter.Hand, card); p.ID != starter.ID && c.KnownBy[p.ID] {
			t.Errorf("a decline revealed the card to %s", p.Name)
		}
	}
}

// ShuffleLibraryKeepingOnTopForEffect leaves the named cards on top, in the
// order given and known to their owner, and shuffles everything else.
func TestShuffleLibraryKeepingOnTopForEffect(t *testing.T) {
	g, _ := revealTable(t)
	p := g.Seats[0]
	lib := p.Library.Cards
	keep := []uuid.UUID{lib[2].InstanceID, lib[0].InstanceID}
	size := len(lib)
	if err := g.ShuffleLibraryKeepingOnTopForEffect(p.ID, keep); err != nil {
		t.Fatal(err)
	}
	if p.Library.Size() != size {
		t.Fatalf("library size %d, want %d", p.Library.Size(), size)
	}
	top := p.Library.Cards[len(p.Library.Cards)-1]
	second := p.Library.Cards[len(p.Library.Cards)-2]
	if top.InstanceID != keep[0] || second.InstanceID != keep[1] {
		t.Fatal("the kept cards are not on top in the order given")
	}
	if !top.KnownBy[p.ID] || !second.KnownBy[p.ID] {
		t.Error("the owner forgot the cards the shuffle left alone")
	}
}
