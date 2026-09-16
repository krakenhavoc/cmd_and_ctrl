package game

import (
	"testing"

	"github.com/google/uuid"
)

// zero_toughness_test.go — #683: the toughness state-based action's
// printed-0 skip (CR 704.5f) no longer covers a printed 0/0 that has
// lost its last counter, and still covers a `*` creature whose 0 is
// the importer's stand-in.
//
// Every road here is one develop already has: the add_counter action
// (AddCounter with a negative delta), an effect removing a counter
// (AddCounterForEffect), and the CR 704.5q +1/+1 / -1/-1 cancel.

// pushZeroZero seats a creature with a printed 0/0 body (Mikaeus,
// Hangarback Walker, an X hydra) holding the given counters.
func pushZeroZero(g *Game, controller *Player, counters map[string]int) uuid.UUID {
	c := NewCard("Zero Zero", controller.ID)
	c.TypeLine = "Artifact Creature — Construct"
	c.Controller = controller.ID
	c.Counters = counters
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// pushStarCreature is a creature whose printed toughness is "*" and
// whose characteristic-defining ability is not coded — Mortivore as
// the importer lands it: Toughness 0, VariableToughness set.
func pushStarCreature(g *Game, controller *Player, counters map[string]int) uuid.UUID {
	c := NewCard("Star Creature", controller.ID)
	c.TypeLine = "Creature — Lhurgoyf"
	c.Controller = controller.ID
	c.VariableToughness = true
	c.Counters = counters
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func runSBAsForTest(g *Game) {
	g.mu.Lock()
	g.stateBasedActionsLocked()
	g.mu.Unlock()
}

// removeLastPlusOne takes one +1/+1 counter off `id` by the named road.
func removeLastPlusOne(t *testing.T, g *Game, id uuid.UUID, road string) {
	t.Helper()
	var err error
	switch road {
	case "add_counter action":
		err = g.AddCounter(id, CounterPlusOne, -1)
	case "effect":
		g.mu.Lock()
		err = g.AddCounterForEffect(id, CounterPlusOne, -1)
		g.mu.Unlock()
	default:
		t.Fatalf("unknown road %q", road)
	}
	if err != nil {
		t.Fatalf("%s: remove counter: %v", road, err)
	}
}

// A printed 0/0 whose LAST +1/+1 counter comes off has a real
// toughness of 0 and is put into its owner's graveyard (CR 704.5f) —
// not skipped as a placeholder because it now has no counters. Before
// Card.LostLastCounter it stayed on the battlefield as a 0/0 no damage
// could kill.
func TestZeroZeroLosingItsLastCounterDies(t *testing.T) {
	for _, road := range []string{"add_counter action", "effect"} {
		t.Run(road, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			walker := pushZeroZero(g, me, map[string]int{CounterPlusOne: 1})
			runSBAsForTest(g)
			if !g.Battlefield.Contains(walker) {
				t.Fatal("a 0/0 holding a +1/+1 counter died before losing it")
			}

			removeLastPlusOne(t, g, walker, road)
			runSBAsForTest(g)
			if g.Battlefield.Contains(walker) {
				t.Fatal("a 0/0 that lost its last +1/+1 counter stayed on the battlefield")
			}
			if !me.Graveyard.Contains(walker) {
				t.Error("the 0/0 did not reach its owner's graveyard")
			}
			for _, c := range me.Graveyard.Cards {
				if c.InstanceID == walker && c.LostLastCounter {
					t.Error("the flag belongs to the permanent; it must not follow the card to the graveyard")
				}
			}
		})
	}
}

// The placeholder convention itself is untouched: a Toughness 0
// creature that never had a counter is still skipped, and so is one
// that still has counters left after losing one.
func TestPrintedZeroSkipStillCoversCreaturesThatNeverLostCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	placeholder := pushZeroZero(g, me, nil)
	big := pushZeroZero(g, me, map[string]int{CounterPlusOne: 2})

	removeLastPlusOne(t, g, big, "add_counter action")
	runSBAsForTest(g)
	if !g.Battlefield.Contains(placeholder) {
		t.Error("a Toughness 0 creature that never had counters was killed — the placeholder skip regressed")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a 0/0 with a +1/+1 counter left died")
	}
	if c := findBattlefieldCard(g, big); c == nil || c.LostLastCounter {
		t.Error("a 0/0 with a counter left was marked LostLastCounter")
	}
}

// CR 704.5q then 704.5f: a 0/0 whose +1/+1 and -1/-1 counters cancel to
// nothing is a 0/0 with no counters, and dies — the same gap, reached
// through the cancel step (Blowfly Infestation on a one-counter
// Hangarback Walker).
func TestCounterCancelToNothingKillsAZeroZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	walker := pushZeroZero(g, me, map[string]int{CounterPlusOne: 1, CounterMinusOne: 1})

	runSBAsForTest(g)
	if g.Battlefield.Contains(walker) {
		t.Error("a 0/0 whose counters cancelled to nothing stayed on the battlefield")
	}
}

// The flag is per object: a permanent that leaves the battlefield sheds
// it with its counters, so the card is a fresh object next time.
func TestLostLastCounterIsClearedLeavingTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = me.ID
	bear.Power, bear.Toughness = 2, 2
	bear.Counters = map[string]int{CounterPlusOne: 1}
	g.Battlefield.PushTop(bear)
	id := bear.InstanceID

	g.mu.Lock()
	_ = g.applyCounterLocked(id, CounterPlusOne, -1)
	g.mu.Unlock()
	if c := findBattlefieldCard(g, id); c == nil || !c.LostLastCounter {
		t.Fatal("removing the last counter did not mark the permanent")
	}
	moved, err := MoveCard(g.Battlefield, me.Hand, id)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if moved.LostLastCounter {
		t.Error("the flag followed the card off the battlefield")
	}
}

// A `*` creature's 0 toughness is the import stand-in, not a printed
// 0, so losing its last counter does not end the placeholder skip: it
// stays on the battlefield by every road a counter comes off — the
// add_counter action, an effect, and the CR 704.5q cancel.
func TestVariableToughnessCreatureSurvivesLosingItsLastCounter(t *testing.T) {
	for _, road := range []string{"add_counter action", "effect"} {
		t.Run(road, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			star := pushStarCreature(g, me, map[string]int{CounterPlusOne: 1})
			removeLastPlusOne(t, g, star, road)
			runSBAsForTest(g)
			if !g.Battlefield.Contains(star) {
				t.Error("a `*` creature died when its last counter was removed")
			}
		})
	}
	t.Run("704.5q cancel", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		star := pushStarCreature(g, me, map[string]int{CounterPlusOne: 1, CounterMinusOne: 1})
		runSBAsForTest(g)
		if !g.Battlefield.Contains(star) {
			t.Error("a `*` creature died when its counters cancelled to nothing")
		}
	})
}

// A copy of a `*` creature copies the stand-in 0, so it copies the bit
// that says so; leaving the battlefield puts the card's own back.
func TestCopyEffectsCarryVariableToughness(t *testing.T) {
	star := NewCard("Star Creature", uuid.New())
	star.TypeLine = "Creature — Lhurgoyf"
	star.VariableToughness = true
	clone := NewCard("Clone", uuid.New())
	clone.TypeLine = "Creature — Shapeshifter"

	clone.applyCopy(CopiableValuesOf(star), star)
	if !clone.VariableToughness {
		t.Error("a copy of a `*` creature lost VariableToughness")
	}
	clone.restorePrintedSelf()
	if clone.VariableToughness {
		t.Error("restoring the copy's own printed values kept VariableToughness")
	}
}

// Removing a counter from a card that has none is not losing its last
// counter: the flag means "went from some to none", so a creature that
// never had a counter is not marked, and a Toughness 0 placeholder a
// misclicked -1 lands on is still skipped by the toughness check.
func TestRemovingACounterFromACardWithNoneDoesNotMarkIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	placeholder := pushZeroZero(g, me, nil)
	if err := g.AddCounter(placeholder, CounterPlusOne, -1); err != nil {
		t.Fatalf("remove counter: %v", err)
	}
	c := findBattlefieldCard(g, placeholder)
	if c == nil {
		t.Fatal("the placeholder left the battlefield")
	}
	if c.LostLastCounter {
		t.Error("removing a counter the card never had marked it LostLastCounter")
	}
	if len(c.Counters) != 0 {
		t.Errorf("counters = %v, want none", c.Counters)
	}
	runSBAsForTest(g)
	if !g.Battlefield.Contains(placeholder) {
		t.Error("a never-countered Toughness 0 creature died after a -1 on nothing")
	}
}
