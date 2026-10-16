package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_placer_test.go pins #3942: every counter placement names the
// player who puts the counters, at the source (counter_placer.go), so
// "if you would put" and "whenever you put" read a fact rather than the
// last resolver in the log.

// placersOn is the Actor of every EventCounterPlaced naming `card`, in
// log order.
func placersOn(g *Game, card uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventCounterPlaced && ev.Target == card {
				out = append(out, ev.Actor)
			}
		}
	})
	return out
}

func placerBear(g *Game, controller uuid.UUID, name string) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: controller, Controller: controller,
	})
}

// CR 608.2c: the controller of a resolving ability follows its
// instructions, so its "put a counter on target creature" is that
// player's placement, whoever controls the creature.
func TestAResolvingAbilityPutsCountersAsItsController(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	theirs := placerBear(g, them, "Their Bear")
	pushAbilityItem(g, StackItemActivated, me, uuid.Nil, "put a counter", func(g *Game, _ *StackItem) error {
		return g.AddCounterForEffect(theirs, CounterMinusOne, 1)
	})
	g.WithWriteLock(func() { g.resolveTopAbilityLocked() })

	if got := placersOn(g, theirs); len(got) != 1 || got[0] != me {
		t.Errorf("placers = %v, want [%v]: the ability's controller puts the counter, not the creature's", got, me)
	}
}

// The issue's own trap. The resolving slot outlives its resolution (it
// is cleared at the next event batch), so a reader that asked "who is
// resolving" after the resolution had ended credited whoever resolved
// last. The rule asks only while the resolution is OPEN, which ends at
// its CR 704.3 boundary.
func TestAPlacementAfterTheResolutionEndedIsNotCreditedToIt(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	mine := placerBear(g, me, "My Bear")
	pushAbilityItem(g, StackItemActivated, them, uuid.Nil, "does nothing", func(*Game, *StackItem) error { return nil })
	g.WithWriteLock(func() { g.resolveTopAbilityLocked() })
	g.RunStateChecksForTest()

	g.WithWriteLock(func() {
		if g.resolving == nil {
			t.Fatal("setup: the resolving slot should still hold the opponent's item until the next batch")
		}
		if err := g.AddCounterForEffect(mine, CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := placersOn(g, mine); len(got) != 1 || got[0] != me {
		t.Errorf("placers = %v, want [%v]: the opponent's resolution had ended", got, me)
	}
}

// A resolution paused on one of its own prompts is restored from a
// snapshot without the resolving slot (it is not persisted); the parked
// tail's item still names the controller for the rest of it.
func TestAPausedResolutionStillNamesItsControllerWithoutTheSlot(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	mine := placerBear(g, me, "My Bear")
	g.WithWriteLock(func() {
		g.resolving = nil
		g.resolutionOpen = true
		g.pausedResolutionTail = &resolutionTail{item: &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: them}}
		if err := g.AddCounterForEffect(mine, CounterMinusOne, 1); err != nil {
			t.Fatal(err)
		}
		g.pausedResolutionTail = nil
		g.resolutionOpen = false
	})
	if got := placersOn(g, mine); len(got) != 1 || got[0] != them {
		t.Errorf("placers = %v, want [%v], the paused resolution's controller", got, them)
	}
}

// CR 122.6a: counters a permanent enters with are put by its controller
// unless the effect says otherwise, so a permanent that enters under my
// control during an opponent's resolution gets them from me. The same
// holds for a planeswalker's starting loyalty (CR 306.5b).
func TestCountersAPermanentEntersWithArePutByItsController(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	mine := placerBear(g, me, "My Bear")
	walker := pushTypedTestCard(g, Card{
		Name: "My Walker", TypeLine: "Legendary Planeswalker — Test", StartingLoyalty: 3,
		Owner: me, Controller: me,
	})
	pushAbilityItem(g, StackItemTriggered, them, uuid.Nil, "puts it onto the battlefield", func(g *Game, _ *StackItem) error {
		g.applyEntryCountersLocked(mine, map[string]int{CounterPlusOne: 2})
		g.stampStartingLoyaltyLocked(walker, "")
		return nil
	})
	g.WithWriteLock(func() { g.resolveTopAbilityLocked() })

	if got := placersOn(g, mine); len(got) != 1 || got[0] != me {
		t.Errorf("entry counter placers = %v, want [%v], the permanent's controller", got, me)
	}
	if got := placersOn(g, walker); len(got) != 1 || got[0] != me {
		t.Errorf("starting loyalty placers = %v, want [%v], the planeswalker's controller", got, me)
	}
}

// An action that is not a resolution names its player for anything it
// places, even inside an open resolution: a mana ability activated to
// pay for something mid-resolution (CR 605.3a) is the activator's.
func TestTheInnermostNamedPlacerWins(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	rock := pushTypedTestCard(g, Card{Name: "My Rock", TypeLine: "Artifact", Owner: me, Controller: me})
	pushAbilityItem(g, StackItemActivated, them, uuid.Nil, "asks for a payment", func(g *Game, _ *StackItem) error {
		pop := g.pushCounterPlacerLocked(me)
		defer pop()
		return g.AddCounterForEffect(rock, "charge", 1)
	})
	g.WithWriteLock(func() { g.resolveTopAbilityLocked() })

	if got := placersOn(g, rock); len(got) != 1 || got[0] != me {
		t.Errorf("placers = %v, want [%v], the activator", got, me)
	}
	g.ReadSnapshot(func() {
		if len(g.counterPlacerScope) != 0 {
			t.Errorf("the scope leaked: %v", g.counterPlacerScope)
		}
	})
}

// CR 714.3c: "that player puts a lore counter on each Saga they
// control" is a turn-based action. It goes through the CR 614 window,
// named and marked, so a replacement that names no effect applies.
func TestASagasLoreCounterIsAPlacementByItsController(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[g.Turn.ActiveSeat].ID
	saga := pushTypedTestCard(g, Card{
		Name: "My Saga", TypeLine: "Enchantment — Saga",
		Owner: active, Controller: active, Counters: map[string]int{CounterLore: 1},
	})
	var seen []ReplacementEvent
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				if ev.Kind == RepEventCounter && ev.CounterTarget == saga {
					seen = append(seen, *ev)
				}
				return false
			},
			Label: "watcher",
		})
		g.advanceSagasForActiveSeatLocked()
	})
	if len(seen) == 0 {
		t.Fatal("the lore counter never opened the CR 614 window")
	}
	if ev := seen[0]; ev.CounterPlacer != active || !ev.CounterFromTurnBasedAction {
		t.Errorf("placer %v, turn-based %v; want %v, true", ev.CounterPlacer, ev.CounterFromTurnBasedAction, active)
	}
	if got := placersOn(g, saga); len(got) != 1 || got[0] != active {
		t.Errorf("placers = %v, want [%v]", got, active)
	}
}
