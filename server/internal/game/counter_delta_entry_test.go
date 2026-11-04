package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_delta_entry_test.go — #4865. An "enters with N counters"
// placement is logged BEFORE the entering zone move and EventETB, so a
// delta walk that stopped at the arrival never saw where the count
// started: a removal from a creature that entered with counters read as
// removing nothing, and a placement read as putting the whole total.
// The walk now stops at the card's last move to a zone other than the
// battlefield instead (CR 400.7), which also keeps an earlier object's
// counters out of a returning card's count.

// entersWithPlusOnes is an entry replacement that gives the card named
// `name` n +1/+1 counters as it enters.
func entersWithPlusOnes(name string, n int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, src *Card) bool {
			if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || src == nil {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.Name == name
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.AddCounterAtETB(CounterPlusOne, n)
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      "enters with counters",
	}
}

// deltaHydraInHand puts a vanilla 0/0 named "Hydra" in seat 0's hand and
// arms the entry replacement that gives it n +1/+1 counters.
func deltaHydraInHand(t *testing.T, g *Game, n int) uuid.UUID {
	t.Helper()
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"hydra-entry": {entersWithPlusOnes("Hydra", n)},
	})
	me := g.Seats[0]
	pushReplacementSource(g, "hydra-entry", me.ID)
	c := NewCard("Hydra", me.ID)
	c.TypeLine = "Creature — Hydra"
	me.Hand.PushTop(c)
	return c.InstanceID
}

// lastCounterDeltas is what the two trigger-side readers say about the
// newest +1/+1 counter event on id.
func lastCounterDeltas(t *testing.T, g *Game, id uuid.UUID) (placed, removed int) {
	t.Helper()
	g.ReadSnapshot(func() {
		for i := len(g.Events) - 1; i >= 0; i-- {
			ev := g.Events[i]
			if ev.Kind == EventCounterPlaced && ev.Target == id && ev.Label == CounterPlusOne {
				placed = g.counterPlacedDeltaLocked(ev, CounterPlusOne)
				removed = g.counterRemovedDeltaLocked(ev, CounterPlusOne)
				return
			}
		}
		t.Fatal("no +1/+1 counter event on the card")
	})
	return placed, removed
}

func deltaAddCounters(t *testing.T, g *Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, CounterPlusOne, n); err != nil {
			t.Fatal(err)
		}
	})
}

// enterHydra puts the Hydra onto the battlefield by one of the two
// entry sites that drain "enters with" counters: the sandbox move and
// an effect's put-onto-the-battlefield.
func enterHydra(t *testing.T, g *Game, me, id uuid.UUID, viaEffect bool) uuid.UUID {
	t.Helper()
	if !viaEffect {
		if err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{Controller: me})
		if err != nil {
			t.Fatal(err)
		}
	})
	return entered
}

func TestCounterDeltasOnAPermanentThatEnteredWithCounters(t *testing.T) {
	for _, via := range []struct {
		name   string
		effect bool
	}{{"sandbox move", false}, {"effect put", true}} {
		t.Run(via.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0].ID
			id := enterHydra(t, g, me, deltaHydraInHand(t, g, 3), via.effect)
			if got := counterCount(g, id, CounterPlusOne); got != 3 {
				t.Fatalf("entered with %d counters, want 3", got)
			}
			if placed, removed := lastCounterDeltas(t, g, id); placed != 3 || removed != 0 {
				t.Errorf("the entry placement: placed %d removed %d, want 3 and 0", placed, removed)
			}

			deltaAddCounters(t, g, id, -1)
			if placed, removed := lastCounterDeltas(t, g, id); placed != 0 || removed != 1 {
				t.Errorf("one removed from 3: placed %d removed %d, want 0 and 1", placed, removed)
			}

			deltaAddCounters(t, g, id, 2)
			if placed, removed := lastCounterDeltas(t, g, id); placed != 2 || removed != 0 {
				t.Errorf("two put on 2: placed %d removed %d, want 2 and 0", placed, removed)
			}
		})
	}
}

// CR 400.7: the card that left and came back is a new object. Its entry
// counters are measured from zero, not from what the old object had, and
// a later removal is measured from what the new object entered with.
func TestCounterDeltasOnACardThatLeftAndCameBackWithCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := enterHydra(t, g, me, deltaHydraInHand(t, g, 2), false)
	deltaAddCounters(t, g, id, 3) // the old object: 5
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneHand, Owner: me}, id); err != nil {
		t.Fatal(err)
	}
	id = enterHydra(t, g, me, id, false)
	if got := counterCount(g, id, CounterPlusOne); got != 2 {
		t.Fatalf("came back with %d counters, want 2", got)
	}
	if placed, removed := lastCounterDeltas(t, g, id); placed != 2 || removed != 0 {
		t.Errorf("the new object's entry: placed %d removed %d, want 2 and 0", placed, removed)
	}
	deltaAddCounters(t, g, id, -2)
	if placed, removed := lastCounterDeltas(t, g, id); placed != 0 || removed != 2 {
		t.Errorf("two removed from the new object's 2: placed %d removed %d, want 0 and 2", placed, removed)
	}
}

// A permanent that entered with none reads as it always did.
func TestCounterDeltasOnAPermanentThatEnteredWithoutCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	id := enterHydra(t, g, me, deltaHydraInHand(t, g, 0), false)
	deltaAddCounters(t, g, id, 2)
	if placed, removed := lastCounterDeltas(t, g, id); placed != 2 || removed != 0 {
		t.Errorf("two put on none: placed %d removed %d, want 2 and 0", placed, removed)
	}
	deltaAddCounters(t, g, id, -1)
	if placed, removed := lastCounterDeltas(t, g, id); placed != 0 || removed != 1 {
		t.Errorf("one removed from 2: placed %d removed %d, want 0 and 1", placed, removed)
	}
}

// counterCount is the number of kind counters on the battlefield card id.
func counterCount(g *Game, id uuid.UUID, kind string) int {
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := g.battlefieldCardLocked(id); ok {
			n = c.Counters[kind]
		}
	})
	return n
}
