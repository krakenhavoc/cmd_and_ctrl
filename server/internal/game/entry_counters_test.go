package game

import (
	"testing"

	"github.com/google/uuid"
)

// entry_counters_test.go — #1002: "this permanent enters with N
// counters on it", N read from the cast (CR 614.1c).
//
// The mechanic, not the thirteen cards that use it — those are pinned
// in cards/effects/entry_counters_cards_test.go. What is here is the
// thing that could not be tested from the catalog at all: that the
// counters are part of the ENTRY EVENT, which is what makes them
// visible to the CR 616 window, to the ETB trigger and to a counter
// payoff, and what the old OnResolve placement could never be.

const (
	testEntryCountersOracle = "test-enters-with-x-counters"
	testEntryKickerOracle   = "test-enters-with-kicked-counters"
	testEntryDoublerOracle  = "test-entry-counter-doubler"
	testEntryScalesOracle   = "test-entry-hardened-scales"
	testEntryWatcherOracle  = "test-entry-counter-watcher"
)

// stubEntryCountersFromCast wires CatalogEntersWithCountersFromCast
// for one oracle ID and restores the previous hook afterwards.
func stubEntryCountersFromCast(t *testing.T, oracleID string, clauses ...EntryCountersFromCast) {
	t.Helper()
	prev := CatalogEntersWithCountersFromCast
	CatalogEntersWithCountersFromCast = func(id string) []EntryCountersFromCast {
		if id == oracleID {
			return clauses
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogEntersWithCountersFromCast = prev })
}

// xEntryCounters is the catalog's XCounters constructor, spelled out
// here so the game package can test the mechanic without importing
// the effects package — escapeCost and flashbackOffer's trick.
func xEntryCounters(kind string) EntryCountersFromCast {
	return EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast CastCounts) int { return cast.X },
	}
}

// kickEntryCounters is the same for CountersPerKick.
func kickEntryCounters(kind string, per int) EntryCountersFromCast {
	return EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast CastCounts) int { return cast.Kicked * per },
	}
}

// seedXCreature puts a printed 0/0 with an {X} cost into the active
// seat's hand and returns its ID. ScryfallID is set because a real
// printing is what tells CR 704.5f's gate that a 0 toughness is a
// real 0 (#691, Card.ToughnessIsKnown).
func seedXCreature(t *testing.T, g *Game, oracleID string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	c := NewCard("Test Hydra", me.ID)
	c.TypeLine = "Creature — Hydra"
	c.ManaCost = "{X}{G}"
	c.OracleID = oracleID
	c.ScryfallID = "test-printing"
	c.Power, c.Toughness = 0, 0
	me.Hand.PushTop(c)
	return c.InstanceID
}

// castX announces the seeded spell for X = n and resolves it.
func castX(t *testing.T, g *Game, id uuid.UUID, n int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(me.ID, id, CastSpellParams{XValue: n}); err != nil {
		t.Fatalf("cast for X=%d: %v", n, err)
	}
	resolveTop(t, g)
}

func counters(t *testing.T, g *Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s is nowhere", id)
	}
	return c.Counters[kind]
}

// The baseline. X counters land on the PERMANENT — they are on it the
// moment it is on the battlefield, and it is the battlefield object
// that carries them, not the card that was on the stack a moment ago.
func TestEntersWithXCountersRidesTheEntry(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	id := seedXCreature(t, g, testEntryCountersOracle)

	castX(t, g, id, 3)

	if !g.Battlefield.Contains(id) {
		t.Fatalf("the X=3 creature did not reach the battlefield")
	}
	if got := counters(t, g, id, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters: got %d, want 3", got)
	}
}

// X=0 is a legal announcement (CR 107.3) and puts no counters on. The
// clause must NOT floor at one: a printed 0/0 cast for X=0 is a 0/0,
// and the next state-based check puts it into its owner's graveyard
// (CR 704.5f, #691) — which is what the counters riding the pipeline
// must not quietly change, since they now land after the move rather
// than before it.
func TestEntersWithXCountersAtZeroPutsNoneAndTheBodyDies(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	id := seedXCreature(t, g, testEntryCountersOracle)

	castX(t, g, id, 0)

	if c, ok := g.LookupCardForEffect(id); ok && len(c.Counters) != 0 {
		t.Errorf("X=0 entered with counters: %v", c.Counters)
	}
	g.WithWriteLock(func() { g.stateBasedActionsLocked() })
	if g.Battlefield.Contains(id) {
		t.Error("a printed 0/0 cast for X=0 survived the toughness check")
	}
}

// CR 107.3b: a permanent that was never a spell has no X. The clause
// is declared on the CARD, so the guard has to be the ENTRY SITE
// rather than the declaration — a Hangarback Walker reanimated, or
// put onto the battlefield from a hand, enters with nothing.
func TestEntersWithXCountersOnlyFromACast(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	id := seedXCreature(t, g, testEntryCountersOracle)
	me := g.Seats[g.Turn.ActiveSeat]

	var entered uuid.UUID
	g.WithWriteLock(func() {
		var err error
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{Controller: me.ID})
		if err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	if entered == uuid.Nil {
		t.Fatal("nothing entered")
	}

	if got := counters(t, g, entered, CounterPlusOne); got != 0 {
		t.Errorf("a permanent put onto the battlefield entered with %d counters, want 0", got)
	}
}

// CR 702.33d, the second thing CastCounts can be read for: "enters
// with a charge counter for each time it was kicked". Kicked once.
func TestEntersWithCountersPerKickOnce(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryKickerOracle, kickEntryCounters("charge", 1))
	stubOptionalCosts(t, testEntryKickerOracle, []AdditionalCost{{
		Key: KickerKey, Label: "Kicker {2}", ManaCost: "{2}", Optional: true,
	}})
	id := seedKickedArtifact(t, g, testEntryKickerOracle)

	castKicked(t, g, id, []int{0})

	if got := counters(t, g, id, "charge"); got != 1 {
		t.Errorf("kicked once: %d charge counters, want 1", got)
	}
}

// Kicked twice — multikicker's count is the LENGTH of the paid list,
// so the same clause scales without a second field.
func TestEntersWithCountersPerKickTwice(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryKickerOracle, kickEntryCounters("charge", 1))
	stubOptionalCosts(t, testEntryKickerOracle, []AdditionalCost{{
		Key: MultikickerKey, Label: "Multikicker {2}", ManaCost: "{2}", Optional: true, Repeat: 3,
	}})
	id := seedKickedArtifact(t, g, testEntryKickerOracle)

	castKicked(t, g, id, []int{0, 0})

	if got := counters(t, g, id, "charge"); got != 2 {
		t.Errorf("kicked twice: %d charge counters, want 2", got)
	}
}

// An unkicked cast of the same card enters with none.
func TestEntersWithCountersPerKickUnkicked(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryKickerOracle, kickEntryCounters("charge", 1))
	stubOptionalCosts(t, testEntryKickerOracle, []AdditionalCost{{
		Key: KickerKey, Label: "Kicker {2}", ManaCost: "{2}", Optional: true,
	}})
	id := seedKickedArtifact(t, g, testEntryKickerOracle)

	castKicked(t, g, id, nil)

	if got := counters(t, g, id, "charge"); got != 0 {
		t.Errorf("unkicked: %d charge counters, want 0", got)
	}
}

// seedKickedArtifact puts a kickable artifact in the active seat's
// hand and advances to a main phase.
func seedKickedArtifact(t *testing.T, g *Game, oracleID string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	c := NewCard("Test Chalice", me.ID)
	c.TypeLine = "Artifact"
	c.ManaCost = "{0}"
	c.OracleID = oracleID
	me.Hand.PushTop(c)
	return c.InstanceID
}

func castKicked(t *testing.T, g *Game, id uuid.UUID, optional []int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(me.ID, id, CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("cast kicked %v: %v", optional, err)
	}
	resolveTop(t, g)
}

// THE POINT OF THE MOVE, half one: Doubling Season. The settled entry
// counters are drained through AddCounterForEffect, which opens the
// ordinary CR 614 counter window, so a doubler applies to them — and
// it applies to a permanent, because by then there is one.
func TestDoublingSeasonDoublesEntryCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		testEntryDoublerOracle: {counterDoubler("Test Doubling Season")},
	})
	pushReplacementSource(g, testEntryDoublerOracle, me.ID)
	id := seedXCreature(t, g, testEntryCountersOracle)

	castX(t, g, id, 3)

	if got := counters(t, g, id, CounterPlusOne); got != 6 {
		t.Errorf("under a doubler: %d counters, want 6", got)
	}
}

// Half two: Hardened Scales, the +1 shape, on the same road.
func TestHardenedScalesAddsOneToEntryCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		testEntryScalesOracle: {counterAdder("Test Hardened Scales", 1)},
	})
	pushReplacementSource(g, testEntryScalesOracle, me.ID)
	id := seedXCreature(t, g, testEntryCountersOracle)

	castX(t, g, id, 3)

	if got := counters(t, g, id, CounterPlusOne); got != 4 {
		t.Errorf("under Hardened Scales: %d counters, want 4", got)
	}
}

// counterAdder is Hardened Scales: one more counter whenever counters
// are put on a permanent its controller controls.
func counterAdder(label string, extra int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterPlaced},
		AppliesTo: func(ev *ReplacementEvent, g *Game, src *Card) bool {
			if ev.Kind != RepEventCounter || src == nil {
				return false
			}
			target, ok := g.LookupCardForEffect(ev.CounterTarget)
			return ok && target.Controller == src.Controller
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.CounterDelta += extra
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      label,
	}
}

// Half three: the card's OWN enters trigger already sees the
// counters, because they are applied after the move and before
// EventETB. Voracious Hydra's "double the number of +1/+1 counters on
// this creature" is the card this is for.
func TestEntryCountersAreThereWhenTheETBTriggerFires(t *testing.T) {
	g := newActiveGame(t)
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	seen := -1
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != testEntryCountersOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventETB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.Kind == EventETB && ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, g *Game) *StackItem {
				seen = source.Counters[CounterPlusOne]
				return nil
			},
		}}
	})
	id := seedXCreature(t, g, testEntryCountersOracle)

	castX(t, g, id, 2)

	if seen != 2 {
		t.Errorf("the enters trigger saw %d counters, want 2", seen)
	}
}

// Half four, and the one the thirteen caveats were about: a
// "whenever one or more counters are put on a permanent you control"
// payoff finally fires, because the counters go on a PERMANENT. The
// old OnResolve put them on a card that was still on the stack, and
// EventCounterPlaced named that card.
func TestEntryCountersFireACounterPayoff(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	stubEntryCountersFromCast(t, testEntryCountersOracle, xEntryCounters(CounterPlusOne))
	fired := 0
	onBattlefield := false
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != testEntryWatcherOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventCounterPlaced},
			AppliesTo: func(ev Event, _ *Card, _ Characteristic, g *Game) bool {
				if ev.Kind != EventCounterPlaced {
					return false
				}
				fired++
				// EventCounterPlaced names the permanent in Target.
				onBattlefield = g.Battlefield.Contains(ev.Target)
				return false
			},
			Build: func(Event, *Card, Characteristic, *Game) *StackItem { return nil },
		}}
	})
	pushReplacementSource(g, testEntryWatcherOracle, me.ID)
	id := seedXCreature(t, g, testEntryCountersOracle)

	castX(t, g, id, 2)

	if fired == 0 {
		t.Fatal("no EventCounterPlaced reached a payoff watching for one")
	}
	if !onBattlefield {
		t.Error("the counters were placed on something that is not a permanent")
	}
}

// The clause and escape's CR 702.138c clause are seeded onto the same
// event one line apart, so a card printing both gets both counts in
// one entry rather than one of them winning.
func TestCastEntryCountersComposeWithEscapeCounters(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-escaping-hydra"
	stubEntryCountersFromCast(t, oracle, xEntryCounters(CounterPlusOne))
	// {X} in the escape cost: CR 107.3b locks X at zero for an
	// alternative cost that does not include it.
	cost := escapeCost("{X}{G}", 2)
	cost.EntersWithCounterName = CounterPlusOne
	cost.EntersWithCounterCount = 1
	withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
	withCatalogAlternativeCosts(t, altCostFor(oracle, cost))
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Escaping Hydra", me.ID)
	c.TypeLine = "Creature — Hydra"
	c.ManaCost = "{X}{G}"
	c.OracleID = oracle
	c.Power, c.Toughness = 0, 0
	me.Graveyard.PushTop(c)
	pay := fodder(me, 2)

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "escape",
		AltCostIDs:      pay,
		XValue:          2,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	resolveTop(t, g)

	if got := counters(t, g, c.InstanceID, CounterPlusOne); got != 3 {
		t.Errorf("X=2 plus escape's one: got %d counters, want 3", got)
	}
}
