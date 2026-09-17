package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestUntapMarkersStayWithObjectsAcrossCopyAndFlicker(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	id := pushTappedPermanent(g, p.ID, "Frozen", "", "Creature", true)
	g.WithWriteLock(func() {
		_ = g.SkipNextUntapForEffect(id, p.ID)
		source := cardByIDForUntapTest(g, id)
		copy := Card{InstanceID: uuid.New(), Owner: p.ID, Controller: p.ID}
		copy.applyCopy(CopiableValuesOf(*source), *source)
		if len(copy.NextUntapSkips) != 0 {
			t.Fatal("copy inherited source's next-untap marker")
		}
		copy.NextUntapSkips = []UntapSkip{{Player: p.ID}}
		copy.applyCopy(CopiableValuesOf(Card{Name: "Another creature"}), Card{})
		if len(copy.NextUntapSkips) != 1 {
			t.Fatal("becoming a copy erased the object's existing marker")
		}
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneExile}, id); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneExile}, ZoneRef{Kind: ZoneBattlefield}, id); err != nil {
		t.Fatal(err)
	}
	if got := cardByIDForUntapTest(g, id).NextUntapSkips; len(got) != 0 {
		t.Fatalf("flickered permanent retained markers: %v", got)
	}
}

func TestOldCardSnapshotRestoresWithoutUntapMarkers(t *testing.T) {
	var snapshot cardSnapshot
	if err := json.Unmarshal([]byte(`{"name":"Old permanent","tapped":true}`), &snapshot); err != nil {
		t.Fatal(err)
	}
	card := restoreCard(&snapshot)
	if !card.Tapped || len(card.NextUntapSkips) != 0 {
		t.Fatalf("old card snapshot restore = %+v", card)
	}
}

func TestStunRemovalHasNoActorAndBypassesCounterReplacement(t *testing.T) {
	g := newActiveGame(t)
	id := pushTappedPermanent(g, g.Seats[0].ID, "Stunned", "", "Artifact", true)
	replacements := 0
	g.WithWriteLock(func() {
		_ = g.applyCounterLocked(id, CounterStun, 2)
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventCounter
			},
			Replace: func(_ *ReplacementEvent, _ *Game, _ *Card) error {
				replacements++
				return nil
			},
		})
	})
	before := len(g.Events)
	if err := g.TapCard(id, false); err != nil {
		t.Fatal(err)
	}
	if c := cardByIDForUntapTest(g, id); !c.Tapped || c.Counters[CounterStun] != 1 {
		t.Fatalf("manual untap did not replace exactly once: %+v", c)
	}
	if replacements != 0 {
		t.Fatal("stun removal entered the counter placement replacement pipeline")
	}
	found := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventUntapCard {
			t.Fatal("replaced untap emitted an untap event")
		}
		if ev.Kind == EventCounterPlaced {
			found = true
			if ev.Actor != uuid.Nil {
				t.Fatal("stun counter removal was attributed to a player")
			}
		}
	}
	if !found {
		t.Fatal("stun removal did not announce its counter change")
	}
}
