package game

import (
	"testing"

	"github.com/google/uuid"
)

// rad_counters_test.go pins #2042: rad counters' inherent triggered
// ability (CR 728.1). At the beginning of a player's precombat main
// phase, if they have rad counters, they mill that many cards; for each
// nonland card milled that way they lose 1 life and remove a rad
// counter. It uses the stack, it has no source and the active player
// controls it, and its life loss is "from radiation" (CR 728.1a).

// radLibrary replaces p's library with `cards`, listed TOP FIRST, so a
// test reads in the order the cards are milled. A name starting with
// "Land" is a basic land; anything else is a nonland creature.
func radLibrary(p *Player, names ...string) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		id := uuid.New()
		c := Card{InstanceID: id, Name: names[i], Owner: p.ID, Controller: p.ID}
		if len(names[i]) >= 4 && names[i][:4] == "Land" {
			c.TypeLine = "Basic Land — Swamp"
		} else {
			c.TypeLine = "Creature — Bear"
			c.Power, c.Toughness = 2, 2
		}
		p.Library.PushTop(c)
		ids[i] = id
	}
	return ids
}

// giveRad puts n rad counters on player through the sandbox verb.
func giveRad(t *testing.T, g *Game, player uuid.UUID, n int) {
	t.Helper()
	if err := g.AddPlayerCounter(player, CounterRad, n); err != nil {
		t.Fatalf("AddPlayerCounter: %v", err)
	}
}

// radTriggersWaiting counts rad triggers queued or on the stack.
func radTriggersWaiting(g *Game) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it != nil && it.Body == radMillBody.Key() {
			n++
		}
	}
	for _, m := range g.StackMeta {
		if m != nil && m.Body == radMillBody.Key() {
			n++
		}
	}
	return n
}

// radLifeEvents returns the EventChangeLife events marked from
// radiation since index `from`.
func radLifeEvents(g *Game, from int) []Event {
	var out []Event
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventChangeLife && ev.FromRadiation {
			out = append(out, ev)
		}
	}
	return out
}

// toPrecombatMain walks the cursor from the opening upkeep into the
// active player's precombat main phase.
func toPrecombatMain(t *testing.T, g *Game) {
	t.Helper()
	advanceIntoStep(t, g, StepPrecombatMain)
}

func TestRadNoCountersNoTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "A", "B", "C")
	toPrecombatMain(t, g)
	if n := radTriggersWaiting(g); n != 0 {
		t.Fatalf("%d rad triggers with no rad counters, want 0", n)
	}
}

// TestRadMillsAndDrainsForNonlandCards is CR 728.1 end to end: three
// rad counters mill three cards, two of them nonland, so the player
// loses 2 life as one loss from radiation and keeps one counter.
func TestRadMillsAndDrainsForNonlandCards(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := radLibrary(me, "Bear", "Land Swamp", "Ogre", "Untouched")
	giveRad(t, g, me.ID, 3)
	life := me.Life

	toPrecombatMain(t, g)
	if n := radTriggersWaiting(g); n != 1 {
		t.Fatalf("%d rad triggers at the precombat main phase, want 1", n)
	}
	// It uses the stack: nothing has happened yet.
	if me.Graveyard.Size() != 0 || me.Life != life {
		t.Fatalf("the trigger acted before resolving: graveyard %d, life %d", me.Graveyard.Size(), me.Life)
	}
	for _, it := range g.PendingTriggers {
		if it.Body == radMillBody.Key() {
			if it.SourceCardID != uuid.Nil {
				t.Errorf("rad trigger has source %v, want none (CR 728.1)", it.SourceCardID)
			}
			if it.Controller != me.ID {
				t.Errorf("rad trigger controlled by %v, want the active player %v", it.Controller, me.ID)
			}
		}
	}

	start := len(g.Events)
	settleMonarchStack(t, g)

	for _, id := range ids[:3] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %v was not milled", id)
		}
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the fourth card was milled; only three rad counters")
	}
	if got := me.Life; got != life-2 {
		t.Errorf("life = %d, want %d (two nonland cards)", got, life-2)
	}
	if got := me.Counters[CounterRad]; got != 1 {
		t.Errorf("rad counters = %d, want 1", got)
	}
	evs := radLifeEvents(g, start)
	if len(evs) != 1 || evs[0].Amount != -2 || evs[0].Target != me.ID {
		t.Errorf("life-from-radiation events = %+v, want one loss of 2 by %v", evs, me.ID)
	}
}

func TestRadOnlyLandsCostsNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "Land A", "Land B")
	giveRad(t, g, me.ID, 2)
	life := me.Life
	toPrecombatMain(t, g)
	settleMonarchStack(t, g)
	if me.Graveyard.Size() != 2 {
		t.Errorf("graveyard %d, want both lands milled", me.Graveyard.Size())
	}
	if me.Life != life || me.Counters[CounterRad] != 2 {
		t.Errorf("life %d rad %d, want %d and 2 — lands cost nothing", me.Life, me.Counters[CounterRad], life)
	}
}

// TestRadLibrarySmallerThanCount: CR 701.17b, mill as many as possible,
// and running out by milling loses nobody the game.
func TestRadLibrarySmallerThanCount(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "Bear", "Ogre")
	giveRad(t, g, me.ID, 5)
	life := me.Life
	toPrecombatMain(t, g)
	settleMonarchStack(t, g)
	g.RunStateChecksForTest()
	if me.Library.Size() != 0 || me.Graveyard.Size() != 2 {
		t.Errorf("library %d graveyard %d, want 0 and 2", me.Library.Size(), me.Graveyard.Size())
	}
	if me.Life != life-2 || me.Counters[CounterRad] != 3 {
		t.Errorf("life %d rad %d, want %d and 3", me.Life, me.Counters[CounterRad], life-2)
	}
	if me.Eliminated {
		t.Error("milling out eliminated the player; only drawing from an empty library does")
	}
}

// TestRadOnlyOnTheirOwnTurn: "each player's precombat main phase" is
// the active player's. An opponent with rad counters is untouched.
func TestRadOnlyOnTheirOwnTurn(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	radLibrary(opp, "Bear")
	giveRad(t, g, opp.ID, 1)
	toPrecombatMain(t, g)
	if n := radTriggersWaiting(g); n != 0 {
		t.Fatalf("%d rad triggers on another player's turn, want 0", n)
	}
}

// TestRadProliferateInResponseMillsMore: the count is read on
// resolution, so a counter added with the trigger on the stack mills
// one more card.
func TestRadProliferateInResponseMillsMore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "A", "B", "C", "D")
	giveRad(t, g, me.ID, 2)
	toPrecombatMain(t, g)
	if radTriggersWaiting(g) != 1 {
		t.Fatal("no rad trigger")
	}
	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, nil, []uuid.UUID{me.ID}); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if me.Counters[CounterRad] != 3 {
		t.Fatalf("rad after proliferate = %d, want 3", me.Counters[CounterRad])
	}
	settleMonarchStack(t, g)
	if me.Graveyard.Size() != 3 {
		t.Errorf("milled %d, want 3 (the count at resolution)", me.Graveyard.Size())
	}
	if me.Counters[CounterRad] != 0 {
		t.Errorf("rad = %d, want 0 (three nonland cards)", me.Counters[CounterRad])
	}
}

// TestRadInterveningIfOnResolution: CR 603.4. Every counter removed in
// response, the ability does nothing.
func TestRadInterveningIfOnResolution(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "A", "B")
	giveRad(t, g, me.ID, 2)
	life := me.Life
	toPrecombatMain(t, g)
	giveRad(t, g, me.ID, -2)
	settleMonarchStack(t, g)
	if me.Graveyard.Size() != 0 || me.Life != life {
		t.Errorf("graveyard %d life %d: the ability acted with no rad counters", me.Graveyard.Size(), me.Life)
	}
}

// TestRadReplacedCardWasNotMilled: a card a replacement sends to exile
// instead never reached the graveyard, so it was not "milled this way"
// and costs no life or counter.
func TestRadReplacedCardWasNotMilled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := radLibrary(me, "Bear", "Ogre")
	giveRad(t, g, me.ID, 2)
	life := me.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(intoGraveyardReplacement(ids[0],
			"if it would be put into a graveyard, exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
	})
	toPrecombatMain(t, g)
	settleMonarchStack(t, g)
	if !g.Exile.Contains(ids[0]) || !me.Graveyard.Contains(ids[1]) {
		t.Fatalf("setup: exile has first %v, graveyard has second %v", g.Exile.Contains(ids[0]), me.Graveyard.Contains(ids[1]))
	}
	if me.Life != life-1 || me.Counters[CounterRad] != 1 {
		t.Errorf("life %d rad %d, want %d and 1: only one card was milled", me.Life, me.Counters[CounterRad], life-1)
	}
}

// TestRadLifeLossIsReplaceableAsRadiation: CR 728.1a. A replacement
// that reads LifeFromRadiation ("you gain life rather than lose life
// from radiation") turns the loss into a gain, and the counters still
// come off.
func TestRadLifeLossIsReplaceableAsRadiation(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "Bear", "Ogre", "Elf")
	giveRad(t, g, me.ID, 3)
	life := me.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife && ev.LifeFromRadiation && ev.LifeDelta < 0
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.LifeDelta = -ev.LifeDelta
				return nil
			},
			Label: "gain life rather than lose life from radiation",
		})
	})
	// An ordinary loss is not from radiation and is not replaced.
	loseLife(t, g, me.ID, 1)
	if me.Life != life-1 {
		t.Fatalf("ordinary loss replaced: life %d", me.Life)
	}
	toPrecombatMain(t, g)
	settleMonarchStack(t, g)
	if me.Life != life-1+3 {
		t.Errorf("life %d, want %d (three gained instead of lost)", me.Life, life+2)
	}
	if me.Counters[CounterRad] != 0 {
		t.Errorf("rad %d, want 0: the counters come off either way", me.Counters[CounterRad])
	}
}

// TestRadTriggerSurvivesARestorePoint: the trigger waiting on the stack
// is a keyed body ("rad/mill"), so a restore point taken then rebuilds
// it and it resolves in the restored game.
func TestRadTriggerSurvivesARestorePoint(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	radLibrary(me, "Bear", "Land", "Ogre")
	giveRad(t, g, me.ID, 3)
	toPrecombatMain(t, g)
	if radTriggersWaiting(g) != 1 {
		t.Fatalf("setup: %d rad triggers", radTriggersWaiting(g))
	}
	if snap := g.CaptureSnapshot(); !snap.Restorable() {
		t.Fatalf("restore point blocked by %v", snap.Continuations.Kinds())
	}
	_, restored := roundTrip(t, g)
	rme := restored.Seats[0]
	life := rme.Life
	if radTriggersWaiting(restored) != 1 {
		t.Fatalf("restored game has %d rad triggers, want 1", radTriggersWaiting(restored))
	}
	settleMonarchStack(t, restored)
	if rme.Graveyard.Size() != 3 || rme.Life != life-2 || rme.Counters[CounterRad] != 1 {
		t.Errorf("restored resolution: graveyard %d life %d rad %d, want 3, %d, 1",
			rme.Graveyard.Size(), rme.Life, rme.Counters[CounterRad], life-2)
	}
}
