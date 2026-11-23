package game

import (
	"testing"

	"github.com/google/uuid"
)

// The cards of an exiled pile land on top of the library, all of them, above
// what was there, and each is a real zone move out of exile.
func TestPutOnTopInRandomOrderFromExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	under := topOfLibraryFor(me, "Under", "Land")
	var pile []uuid.UUID
	for i := 0; i < 4; i++ {
		c := NewCard("Exiled", me.ID)
		c.TypeLine = "Instant"
		g.Exile.PushTop(c)
		pile = append(pile, c.InstanceID)
	}
	before := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.PutOnTopInRandomOrderForEffect(me.ID, ZoneExile, pile); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.Exile.Cards) != 0 {
		t.Fatalf("%d cards stayed in exile", len(g.Exile.Cards))
	}
	cards := me.Library.Cards
	if len(cards) < 5 {
		t.Fatalf("library has %d cards, want the old top plus the pile", len(cards))
	}
	top := map[uuid.UUID]bool{}
	for _, c := range cards[len(cards)-4:] {
		top[c.InstanceID] = true
	}
	for _, id := range pile {
		if !top[id] {
			t.Errorf("card %s is not among the top four", id)
		}
	}
	if cards[len(cards)-5].InstanceID != under {
		t.Error("the old top card is not directly under the pile")
	}
	moves := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventZoneMove && ev.OldZone == ZoneExile && ev.NewZone == ZoneLibrary {
			moves++
		}
	}
	if moves != 4 {
		t.Errorf("%d zone moves, want 4", moves)
	}
}

// A card that is no longer in the named zone is left alone, and a repeated id
// is moved once.
func TestPutOnTopInRandomOrderSkipsCardsOutOfTheZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	in := NewCard("In Exile", me.ID)
	in.TypeLine = "Instant"
	g.Exile.PushTop(in)
	elsewhere := NewCard("In Hand", me.ID)
	elsewhere.TypeLine = "Instant"
	me.Hand.PushTop(elsewhere)
	size := me.Library.Size()
	g.WithWriteLock(func() {
		if err := g.PutOnTopInRandomOrderForEffect(me.ID, ZoneExile, []uuid.UUID{in.InstanceID, in.InstanceID, elsewhere.InstanceID}); err != nil {
			t.Fatal(err)
		}
	})
	if me.Library.Size() != size+1 {
		t.Errorf("library grew by %d, want 1", me.Library.Size()-size)
	}
	if !me.Hand.Contains(elsewhere.InstanceID) {
		t.Error("a card from another zone was moved")
	}
}
