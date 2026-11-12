package game

import (
	"testing"

	"github.com/google/uuid"
)

// chaos_ensues_test.go — #5648: "chaos ensues" said by an effect (CR
// 311.7) and the planar flag on a dice roll's window.

// CR 311.7: an effect's "chaos ensues" triggers a face-up plane's chaos
// ability, controlled by the planar controller whoever controls the
// effect, with no die rolled; a plane in the deck triggers nothing.
func TestChaosEnsuesForEffectTriggersAFaceUpPlanesChaosAbility(t *testing.T) {
	g, p := planechaseMain(t)
	other := g.Seats[1-g.Turn.ActiveSeat]
	plane, deckPlane := g.PlanarFaceUpCards()[0], g.PlanarDeckCards()[0]
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != plane.OracleID && key != deckPlane.OracleID {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventRollDie, EventChaosEnsues}, Key: "chaos",
			AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool {
				return ev.Kind == EventChaosEnsues || ev.PlanarFace == PlanarFaceChaos
			},
			Effect: func(*Game, *StackItem) error { return nil },
		}}
	})
	source := uuid.New()
	g.WithWriteLock(func() { g.ChaosEnsuesForEffect(other.ID, source) })
	if countSourced(g, plane.InstanceID) != 1 {
		t.Fatal("chaos from an effect did not trigger the face-up plane's ability")
	}
	for _, it := range g.PendingTriggers {
		if it.SourceCardID == plane.InstanceID && it.Controller != p.ID {
			t.Errorf("the chaos ability is controlled by %v, want the planar controller", it.Controller)
		}
	}
	if countSourced(g, deckPlane.InstanceID) != 0 {
		t.Error("a face-down plane's ability triggered")
	}
	if _, rolled := lastRoll(g); rolled {
		t.Error("chaos from an effect announced a die roll")
	}
	announced := 0
	for _, ev := range g.Events {
		if ev.Kind == EventChaosEnsues {
			announced++
			if ev.Actor != other.ID || ev.Source != source {
				t.Errorf("the event is %v from %v, want the effect's controller and source", ev.Actor, ev.Source)
			}
		}
	}
	if announced != 1 {
		t.Errorf("%d chaos-ensues events, want 1", announced)
	}
	// The die's chaos face still triggers the same ability.
	forcePlanarFace(t, g, p.ID, 1)
	if countSourced(g, plane.InstanceID) != 2 {
		t.Error("the planar die's chaos face no longer triggers the ability")
	}
}

// With no planar deck there is no chaos ability: nothing is announced.
func TestChaosEnsuesForEffectDoesNothingWithoutPlanechase(t *testing.T) {
	g := newActiveGame(t)
	n := len(g.Events)
	g.WithWriteLock(func() { g.ChaosEnsuesForEffect(g.Seats[0].ID, uuid.Nil) })
	if len(g.Events) != n {
		t.Errorf("chaos ensued at a table without Planechase: %+v", g.Events[n:])
	}
}

// RollOneMoreAndIgnoreOne adds a die and a chooser to a dice roll's
// window, and does nothing to any other event.
func TestRollOneMoreAndIgnoreOne(t *testing.T) {
	chooser, source := uuid.New(), uuid.New()
	ev := &ReplacementEvent{Kind: RepEventRollDice, RollCount: 1, RollPlanar: true}
	ev.RollOneMoreAndIgnoreOne(chooser, source, "first")
	ev.RollOneMoreAndIgnoreOne(chooser, source, "second")
	if ev.RollCount != 3 || len(ev.rollShields) != 2 || ev.rollShields[0].Label != "first" ||
		ev.rollShields[1].Chooser != chooser || ev.rollShields[1].Source != source {
		t.Errorf("after two replacements: count %d, shields %+v", ev.RollCount, ev.rollShields)
	}
	other := &ReplacementEvent{Kind: RepEventDamage, RollCount: 1}
	other.RollOneMoreAndIgnoreOne(chooser, source, "x")
	none := &ReplacementEvent{Kind: RepEventRollDice}
	none.RollOneMoreAndIgnoreOne(chooser, source, "x")
	if other.RollCount != 1 || len(other.rollShields) != 0 || none.RollCount != 0 || len(none.rollShields) != 0 {
		t.Error("an event that is not a roll of one or more dice was changed")
	}
}
