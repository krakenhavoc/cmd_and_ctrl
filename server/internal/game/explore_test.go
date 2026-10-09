package game

import (
	"testing"

	"github.com/google/uuid"
)

// explore_test.go — #2720, CR 701.44: "<permanent> explores".

// exploreSetup puts a creature under seat 0 and `top` on top of its
// library, and returns the creature's ObjectRef.
func exploreSetup(t *testing.T, top Card) (*Game, *Player, ObjectRef) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushColouredCreature(g, me, "Explorer", []string{"G"})
	top.Owner, top.Controller = me.ID, me.ID
	me.Library.PushTop(top)
	c, _ := battlefieldCardByID(g, id)
	return g, me, ObjectRef{ID: id, Epoch: c.ObjectEpoch}
}

func countExplored(g *Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventExplored && ev.CardID == id {
			n++
		}
	}
	return n
}

func openConfirm(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceConfirm {
			return c
		}
	}
	return nil
}

// TestExploreRevealsALandIntoHand — a land goes to hand, no counter, no
// question, and the permanent has explored.
func TestExploreRevealsALandIntoHand(t *testing.T) {
	land := NewCard("Forest", uuid.Nil)
	land.TypeLine = "Basic Land — Forest"
	g, me, explorer := exploreSetup(t, land)
	hand := me.Hand.Size()

	g.WithWriteLock(func() {
		if err := g.ExploreForEffect(uuid.Nil, explorer, me.ID, nil); err != nil {
			t.Fatalf("ExploreForEffect: %v", err)
		}
	})
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(land.InstanceID) {
		t.Error("the revealed land did not go to hand")
	}
	if c, _ := battlefieldCardByID(g, explorer.ID); c.Counters[CounterPlusOne] != 0 {
		t.Error("a land explore put a counter on the explorer")
	}
	if openConfirm(g) != nil {
		t.Error("a land explore asked the graveyard question")
	}
	if countExplored(g, explorer.ID) != 1 {
		t.Error("no explored event")
	}
}

// TestExploreNonlandCountersThenAsks — a nonland puts a +1/+1 counter on
// the explorer and asks; yes puts the card into the graveyard, and the
// explored event follows the answer.
func TestExploreNonlandCountersThenAsks(t *testing.T) {
	for _, accept := range []bool{true, false} {
		spell := NewCard("Some Sorcery", uuid.Nil)
		spell.TypeLine = "Sorcery"
		g, me, explorer := exploreSetup(t, spell)

		g.WithWriteLock(func() {
			if err := g.ExploreForEffect(uuid.Nil, explorer, me.ID, nil); err != nil {
				t.Fatalf("ExploreForEffect: %v", err)
			}
		})
		if c, _ := battlefieldCardByID(g, explorer.ID); c.Counters[CounterPlusOne] != 1 {
			t.Errorf("explorer has %d +1/+1 counters, want 1", c.Counters[CounterPlusOne])
		}
		q := openConfirm(g)
		if q == nil {
			t.Fatal("no graveyard question")
		}
		if countExplored(g, explorer.ID) != 0 {
			t.Error("explored before the question was answered")
		}
		if err := g.ResolveConfirm(q.ID, me.ID, accept); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		inYard := me.Graveyard.Contains(spell.InstanceID)
		onTop := me.Library.Size() > 0 && me.Library.Cards[me.Library.Size()-1].InstanceID == spell.InstanceID
		if accept && !inYard {
			t.Error("yes: the card is not in the graveyard")
		}
		if !accept && !onTop {
			t.Error("no: the card is not still on top")
		}
		if countExplored(g, explorer.ID) != 1 {
			t.Errorf("accept=%v: explored %d times, want 1", accept, countExplored(g, explorer.ID))
		}
	}
}

// TestExploreByADepartedPermanentStillReveals — CR 701.44c and 400.7: a
// stale ObjectRef gets no counter, but the reveal and the question
// still happen.
func TestExploreByADepartedPermanentStillReveals(t *testing.T) {
	spell := NewCard("Some Sorcery", uuid.Nil)
	spell.TypeLine = "Sorcery"
	g, me, explorer := exploreSetup(t, spell)
	stale := ObjectRef{ID: explorer.ID, Epoch: explorer.Epoch + 1}

	g.WithWriteLock(func() {
		if err := g.ExploreForEffect(uuid.Nil, stale, me.ID, nil); err != nil {
			t.Fatalf("ExploreForEffect: %v", err)
		}
	})
	if c, _ := battlefieldCardByID(g, explorer.ID); c.Counters[CounterPlusOne] != 0 {
		t.Error("a different object got the counter")
	}
	if openConfirm(g) == nil {
		t.Error("the graveyard question was not asked")
	}
}

// TestExploreWithAnEmptyLibraryStillExplores — CR 701.44b.
func TestExploreWithAnEmptyLibraryStillExplores(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushColouredCreature(g, me, "Explorer", []string{"G"})
	me.Library.Cards = nil
	c, _ := battlefieldCardByID(g, id)
	g.WithWriteLock(func() {
		if err := g.ExploreForEffect(uuid.Nil, ObjectRef{ID: id, Epoch: c.ObjectEpoch}, me.ID, nil); err != nil {
			t.Fatalf("ExploreForEffect: %v", err)
		}
	})
	if countExplored(g, id) != 1 {
		t.Error("an explore with nothing to reveal did not count as explored")
	}
}
