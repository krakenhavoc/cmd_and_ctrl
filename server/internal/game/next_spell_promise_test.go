package game

import (
	"testing"

	"github.com/google/uuid"
)

// next_spell_promise_test.go — #1852: "the next <kind> spell you cast
// this turn <does X>". The engine contract with hand-built spells; the
// printed cards are pinned in cards/effects/next_spell_promise_test.go
// and the enumerator's agreement in legal/next_spell_promise_test.go.

func grantPromise(g *Game, p *Player, np NextSpellPromise, label string) {
	g.WithWriteLock(func() {
		g.GrantNextSpellPromiseForEffect(p.ID, np, label, uuid.New(), Duration{})
	})
}

func clearStack(g *Game) {
	g.WithWriteLock(func() {
		g.Stack.Cards = nil
		g.StackMeta = nil
	})
}

func promisePriceOf(t *testing.T, g *Game, p *Player, card Card, cost string) ParsedCost {
	t.Helper()
	base, err := ParseCost(cost)
	if err != nil {
		t.Fatal(err)
	}
	var out ParsedCost
	g.ReadSnapshot(func() {
		out, err = g.ApplyCostModifiersForEffect(base, CostQuery{Card: card, Controller: p.ID, FromZone: ZoneHand})
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func promisesOn(p *Player) int {
	n := 0
	for _, s := range p.Statics {
		if s.NextSpell.Active {
			n++
		}
	}
	return n
}

// The promise is spent by the first spell its filter matches, and only
// by it; a spell that does not match leaves it for the next one, and
// another player's casts never touch it.
func TestNextSpellPromiseIsSpentByTheFirstMatchingSpellOnly(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	grantPromise(g, me, NextSpellPromise{
		Filter: PermissionFilter{CreatureOnly: true}, CantBeCountered: true,
		CounterKind: "+1/+1", Counters: 1, Text: "next creature",
	}, "Savage Summoning")

	sorcery := castFromHand(t, g, me, "Divination", "Sorcery")
	if promisesOn(me) != 1 || len(g.StackMeta[sorcery].PromisedCounters) != 0 {
		t.Fatalf("a sorcery does not match a creature promise: %+v", me.Statics)
	}
	clearStack(g)

	// An opponent's creature spell is not mine to spend.
	theirs := pushSpellFor(t, g, "Creature — Bear", opp, opp)
	g.WithWriteLock(func() { g.spendNextSpellPromisesLocked(opp.ID, theirs) })
	if promisesOn(me) != 1 || len(g.StackMeta[theirs].PromisedCounters) != 0 {
		t.Fatal("only the promise's own player spends it")
	}
	clearStack(g)

	bear := castFromHand(t, g, me, "Grizzly Bears", "Creature — Bear")
	if promisesOn(me) != 0 {
		t.Fatal("the first matching spell spends the promise")
	}
	marks := g.StackMeta[bear].PromisedCounters
	if len(marks) != 1 || marks[0].Kind != "+1/+1" || marks[0].Count != 1 || marks[0].SourceName != "Savage Summoning" {
		t.Errorf("the counter mark: %+v", marks)
	}
	if !cantBeCountered(g, bear) {
		t.Error("the uncounterable rider became a mark on the spell")
	}
	clearStack(g)

	second := castFromHand(t, g, me, "Second Bear", "Creature — Bear")
	if len(g.StackMeta[second].PromisedCounters) != 0 || cantBeCountered(g, second) {
		t.Error("a promise is spent once")
	}
}

// A spell that was cast and then countered still used the promise: it
// is spent at CR 601.2i, not at resolution.
func TestACounteredSpellStillSpentThePromise(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantPromise(g, me, NextSpellPromise{Reduce: 1, Text: "cheaper"}, "Berserker")
	bear := castFromHand(t, g, me, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if g.Stack.Contains(bear) {
		t.Fatal("setup: the spell should have been countered")
	}
	if promisesOn(me) != 0 {
		t.Error("the countered spell still consumed the promise")
	}
}

// "This turn": the promise ends at cleanup (CR 514.2); one with no
// stated end (Xho Cai) waits.
func TestNextSpellPromiseExpiresAtEndOfTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := NewCard("Test Bear", me.ID)
	card.TypeLine, card.ManaCost = "Creature — Bear", "{2}{G}"
	grantPromise(g, me, NextSpellPromise{Reduce: 1, Text: "this turn"}, "A")
	g.WithWriteLock(func() {
		g.GrantNextSpellPromiseForEffect(me.ID, NextSpellPromise{Reduce: 1, Text: "no end"}, "B", uuid.New(), IndefiniteDuration())
	})
	if got := promisePriceOf(t, g, me, card, "{2}{G}"); got.Generic != 0 {
		t.Fatalf("both promises apply while live: %+v", got)
	}
	endTheTurn(g)
	if promisesOn(me) != 1 {
		t.Fatalf("the this-turn promise ends at cleanup and the open-ended one stays: %+v", me.Statics)
	}
	if got := promisePriceOf(t, g, me, card, "{2}{G}"); got.Generic != 1 {
		t.Errorf("only the surviving promise prices the spell: %+v", got)
	}
}

// Flash is read before the cast, folded into the one timing read, and
// is spent by the cast it permits — even a cast that did not need it.
func TestFlashPromiseOpensTheWindowForMatchingSpellsOnly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	bear := timingHandCard(me, "Test Bear", "Creature — Bear")
	sorcery := timingHandCard(me, "Test Sorcery", "Sorcery")
	if timingOpenFor(t, g, me, bear) {
		t.Fatal("setup: a creature spell is not castable in an upkeep step")
	}
	grantPromise(g, me, NextSpellPromise{Filter: PermissionFilter{CreatureOnly: true}, Flash: true, Text: "flash"}, "Savage Summoning")
	if !timingOpenFor(t, g, me, bear) {
		t.Fatal("the promise opens instant speed for a creature spell")
	}
	if timingOpenFor(t, g, me, sorcery) {
		t.Fatal("and not for a sorcery")
	}
	if err := g.CastSpell(me.ID, bear, CastSpellParams{}); err != nil {
		t.Fatalf("cast under the promise: %v", err)
	}
	if promisesOn(me) != 0 {
		t.Error("the cast it permitted spent it")
	}
	again := timingHandCard(me, "Second Bear", "Creature — Bear")
	if timingOpenFor(t, g, me, again) {
		t.Error("the window closes once the promise is spent")
	}
}

// A promise used at sorcery speed is spent all the same: the rulings
// say the cast uses it up whether or not it needed the flash.
func TestFlashPromiseIsSpentEvenWhenCastAtSorcerySpeed(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantPromise(g, me, NextSpellPromise{Filter: PermissionFilter{CreatureOnly: true}, Flash: true}, "Savage Summoning")
	castFromHand(t, g, me, "Grizzly Bears", "Creature — Bear")
	if promisesOn(me) != 0 {
		t.Error("a main-phase cast spends the flash promise")
	}
}

// The cost rider prices the cast in CR 601.2f, floors at the generic
// component, stacks with a second promise, ignores a spell its filter
// does not match, and is not consumed by pricing.
func TestCostPromiseReducesGenericOnly(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := NewCard("Test Bear", me.ID)
	bear.TypeLine, bear.ManaCost = "Creature — Bear", "{1}{G}"
	bolt := NewCard("Test Bolt", me.ID)
	bolt.TypeLine, bolt.ManaCost = "Instant", "{1}{R}"

	grantPromise(g, me, NextSpellPromise{Filter: PermissionFilter{InstantOrSorceryOnly: true}, Reduce: 3, Text: "X less"}, "Kaza")
	if got := promisePriceOf(t, g, me, bear, "{1}{G}"); got.Generic != 1 {
		t.Errorf("a creature is not an instant or sorcery: %+v", got)
	}
	got := promisePriceOf(t, g, me, bolt, "{1}{R}")
	if got.Generic != 0 || got.ManaValue() != 1 {
		t.Errorf("a reduction never touches a coloured requirement: %+v", got)
	}
	if promisePriceOf(t, g, me, bolt, "{1}{R}"); promisesOn(me) != 1 {
		t.Error("pricing a cast does not spend the promise")
	}
	grantPromise(g, me, NextSpellPromise{Reduce: 1}, "Berserker")
	if got := promisePriceOf(t, g, me, bear, "{3}{G}"); got.Generic != 2 {
		t.Errorf("only the unfiltered promise applies to a creature: %+v", got)
	}
}

// The extra counter rides the entry pipeline: the creature spell's
// stack item carries the mark, and the permanent arrives with it.
func TestPromisedCounterLandsOnTheEnteringCreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantPromise(g, me, NextSpellPromise{Filter: PermissionFilter{CreatureOnly: true}, CounterKind: "+1/+1", Counters: 2}, "Test")
	bear := castFromHand(t, g, me, "Grizzly Bears", "Creature — Bear")
	for i := 0; i < 8 && g.Stack.Contains(bear); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	var counters int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Grizzly Bears" {
				counters = c.Counters["+1/+1"]
			}
		}
	})
	if counters != 2 {
		t.Errorf("the creature entered with %d +1/+1 counters, want 2", counters)
	}
}

// A copy is not cast (CR 707.10), so it neither spends the promise nor
// receives its marks.
func TestACopyNeverSpendsANextSpellPromise(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	grantPromise(g, me, NextSpellPromise{CantBeCountered: true, Reduce: 1}, "Test")
	spell := pushSpellFor(t, g, "Instant", me, me)
	g.WithWriteLock(func() { g.StackMeta[spell].IsCopy = true })
	var keys []CastFollowUpKey
	g.WithWriteLock(func() { keys = g.spendNextSpellPromisesLocked(me.ID, spell) })
	if keys != nil || promisesOn(me) != 1 || cantBeCountered(g, spell) {
		t.Error("a copy is not cast and spends nothing")
	}
}

// A promise whose rider is a registered follow-up runs it once the
// spell is on the stack — the RegisterCastFollowUp seam, reused.
func TestNextSpellPromiseRunsItsRegisteredFollowUp(t *testing.T) {
	var ran []CastFollowUp
	key := RegisterCastFollowUp("test-1852-follow-up", func(_ *Game, f CastFollowUp) error {
		ran = append(ran, f)
		return nil
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantPromise(g, me, NextSpellPromise{FollowUp: key}, "Test")
	bear := castFromHand(t, g, me, "Grizzly Bears", "Creature — Bear")
	if len(ran) != 1 || ran[0].Player != me.ID || ran[0].Spell != bear {
		t.Fatalf("the follow-up ran %+v", ran)
	}
	clearStack(g)
	castFromHand(t, g, me, "Second Bear", "Creature — Bear")
	if len(ran) != 1 {
		t.Error("it runs once")
	}
}

// The promise and the mark are plain data and survive a restore point.
func TestNextSpellPromiseAndMarkSurviveARestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantPromise(g, me, NextSpellPromise{Filter: PermissionFilter{CreatureOnly: true}, Flash: true, Reduce: 1, Text: "live"}, "Live")
	grantPromise(g, me, NextSpellPromise{Filter: PermissionFilter{SorceryOnly: true}, CounterKind: "+1/+1", Counters: 1}, "Spent")
	sorcery := castFromHand(t, g, me, "Divination", "Sorcery")

	_, restored := roundTrip(t, g)
	rme := restored.playerByIDLocked(me.ID)
	if len(rme.Statics) != 1 || rme.Statics[0].NextSpell != me.Statics[0].NextSpell {
		t.Fatalf("the unspent promise: %+v", rme.Statics)
	}
	got := restored.StackMeta[sorcery].PromisedCounters
	if len(got) != 1 || got[0] != g.StackMeta[sorcery].PromisedCounters[0] {
		t.Errorf("the spent promise's mark: %+v", got)
	}
	card := NewCard("Test Bear", rme.ID)
	card.TypeLine, card.ManaCost = "Creature — Bear", "{1}{G}"
	if got := promisePriceOf(t, restored, rme, card, "{1}{G}"); got.Generic != 0 {
		t.Errorf("the restored promise still prices a creature spell: %+v", got)
	}
}
