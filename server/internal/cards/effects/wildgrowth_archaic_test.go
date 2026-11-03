package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wildgrowth_archaic_test.go — #4686: a cast trigger that gives the
// creature spell it watched "enters with X additional +1/+1 counters",
// played through the real catalog entry and the real stack.

const wildgrowthArchaicOracle = "4469ff35-54ec-4ff5-bc19-3808ae0f711b"

// wildgrowthOnBattlefield seeds an Archaic for `owner`. Seeded rather
// than cast, so it is given toughness to survive without its converge
// counters; the converge clause has its own test.
func wildgrowthOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Wildgrowth Archaic", "Creature — Avatar", wildgrowthArchaicOracle, 1, 1)
}

// wildgrowthCastBear puts a {1}{G} 2/2 in `p`'s hand, floats `mana`
// and casts it, returning its ID. The cast trigger, if any, is then on
// the stack above it.
func wildgrowthCastBear(t *testing.T, g *game.Game, p *game.Player, mana string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	})
	floatForTest(g, p, mana)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	return id
}

// wildgrowthTriggerOnStack is the Archaic's cast trigger waiting on the
// stack, or uuid.Nil.
func wildgrowthTriggerOnStack(t *testing.T, g *game.Game, archaic uuid.UUID) uuid.UUID {
	t.Helper()
	var out uuid.UUID
	g.WithWriteLock(func() {
		for id, item := range g.StackMeta {
			if item.Kind != game.StackItemSpell && item.SourceCardID == archaic {
				out = id
			}
		}
	})
	return out
}

func TestWildgrowthArchaicGivesTheCreatureAnExtraCounterPerColourSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	archaic := wildgrowthOnBattlefield(g, me.ID)
	bear := wildgrowthCastBear(t, g, me, "GU")
	if wildgrowthTriggerOnStack(t, g, archaic) == uuid.Nil {
		t.Fatal("casting a creature spell did not trigger the Archaic")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear entered with %d +1/+1 counters, want 2 (green and blue were spent)", got)
	}
	if got := counterCount(g, archaic, game.CounterPlusOne); got != 0 {
		t.Errorf("the Archaic has %d counters; the counters are the creature's", got)
	}
}

func TestWildgrowthArchaicCountsColoursNotMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	wildgrowthOnBattlefield(g, me.ID)
	mono := wildgrowthCastBear(t, g, me, "GG")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mono, game.CounterPlusOne); got != 1 {
		t.Errorf("two green is one colour: the Bear has %d counters, want 1", got)
	}
	colourless := wildgrowthCastBear(t, g, me, "GC")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, colourless, game.CounterPlusOne); got != 1 {
		t.Errorf("colourless mana is not a colour: the Bear has %d counters, want 1", got)
	}
}

func TestWildgrowthArchaicCounteredTriggerGivesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	archaic := wildgrowthOnBattlefield(g, me.ID)
	bear := wildgrowthCastBear(t, g, me, "GU")
	trig := wildgrowthTriggerOnStack(t, g, archaic)
	if trig == uuid.Nil {
		t.Fatal("no trigger on the stack")
	}
	var err error
	g.WithWriteLock(func() { err = g.CounterTargetForEffect(trig) })
	if err != nil {
		t.Fatalf("counter the trigger: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, bear) {
		t.Fatal("the Bear did not resolve")
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the trigger was countered, yet the Bear has %d counters", got)
	}
}

func TestWildgrowthArchaicCounteredCreatureGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	archaic := wildgrowthOnBattlefield(g, me.ID)
	bear := wildgrowthCastBear(t, g, me, "GU")
	trig := wildgrowthTriggerOnStack(t, g, archaic)
	var err error
	g.WithWriteLock(func() { err = g.CounterTargetForEffect(bear) })
	if err != nil {
		t.Fatalf("counter the Bear: %v", err)
	}
	g.WithWriteLock(func() {
		if item := g.StackMeta[trig]; item == nil || item.Params.Amount != 2 {
			t.Errorf("the trigger should still be waiting with X = 2: %+v", item)
		}
	})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, bear) {
		t.Fatal("a countered Bear is on the battlefield")
	}
	if got := counterCount(g, archaic, game.CounterPlusOne); got != 0 {
		t.Errorf("the counters went to the Archaic (%d)", got)
	}
	// The mark refuses the card now that it is not a spell on the stack.
	g.WithWriteLock(func() {
		if g.MarkSpellEntersWithCountersForEffect(game.ObjectRef{ID: bear, Epoch: -1}, game.PromisedCounter{Kind: game.CounterPlusOne, Count: 2}) {
			t.Error("a card in the graveyard was marked as a spell on the stack")
		}
	})
}

func TestWildgrowthArchaicConvergeStacksWithTheExtraCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	wildgrowthOnBattlefield(g, me.ID)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wildgrowth Archaic", TypeLine: "Creature — Avatar",
			ManaCost: "{2/G}{2/G}", OracleID: wildgrowthArchaicOracle, Owner: me.ID, Controller: me.ID})
	})
	// {G} for one hybrid symbol, {2} paid with blue for the other: two
	// colours.
	floatForTest(g, me, "GUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the second Archaic: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatal("the second Archaic is not on the battlefield")
	}
	if got := counterCount(g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("the second Archaic has %d +1/+1 counters, want 4 (converge 2 + the first's 2)", got)
	}
}

func TestWildgrowthArchaicConvergeAloneAndForNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	cast := func(mana string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wildgrowth Archaic", TypeLine: "Creature — Avatar",
				ManaCost: "{2/G}{2/G}", OracleID: wildgrowthArchaicOracle, Owner: me.ID, Controller: me.ID,
				PrintedPTKnown: true})
		})
		floatForTest(g, me, mana)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast with %s: %v", mana, err)
		}
		passPriorityAroundTable(t, g)
		return id
	}
	gg := cast("GG")
	if got := counterCount(g, gg, game.CounterPlusOne); got != 1 {
		t.Errorf("cast with {G}{G}: %d counters, want 1 (one colour)", got)
	}
	// The first one is on the battlefield now; it gives the second its
	// own count again, so the generic-paid one still has none from
	// converge and none from the trigger.
	generic := cast("CCCC")
	if onBattlefield(g, generic) {
		t.Errorf("cast for four colourless it is a 0/0 with %d counters and should have died",
			counterCount(g, generic, game.CounterPlusOne))
	}
}

func TestWildgrowthArchaicCopyOfTheCreatureSpellGetsNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	wildgrowthOnBattlefield(g, me.ID)
	bear := wildgrowthCastBear(t, g, me, "GU")
	var copyID uuid.UUID
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(bear, me.ID, false, nil); err != nil {
			t.Fatalf("copy the Bear: %v", err)
		}
		for id, item := range g.StackMeta {
			if item.IsCopy {
				copyID = id
			}
		}
		if copyID == uuid.Nil {
			t.Fatal("no copy on the stack")
		}
		if g.MarkSpellEntersWithCountersForEffect(game.ObjectRef{ID: copyID, Epoch: g.ObjectEpochForEffect(copyID)},
			game.PromisedCounter{Kind: game.CounterPlusOne, Count: 1}) {
			t.Error("a copy of a spell was marked")
		}
	})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the cast Bear has %d counters, want 2", got)
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Bear" && c.InstanceID != bear {
			token = c
		}
	}
	if token == nil {
		t.Fatal("the copy did not become a token on the battlefield")
	}
	if got := token.Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("the token copy of the Bear has %d counters; it was not cast and was not marked", got)
	}
}

func TestWildgrowthArchaicIgnoresOpponentsAndNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	archaic := wildgrowthOnBattlefield(g, opp.ID)
	bear := wildgrowthCastBear(t, g, me, "GU")
	if wildgrowthTriggerOnStack(t, g, archaic) != uuid.Nil {
		t.Error("an opponent's creature spell triggered the Archaic")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the Bear has %d counters from an opponent's Archaic", got)
	}

	mine := wildgrowthOnBattlefield(g, me.ID)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}{G}",
			Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "GU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	if wildgrowthTriggerOnStack(t, g, mine) != uuid.Nil {
		t.Error("a noncreature spell triggered the Archaic")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("the artifact has %d counters", got)
	}
}

func TestWildgrowthArchaicCountersAreSeenByHardenedScales(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	wildgrowthOnBattlefield(g, me.ID)
	bear := wildgrowthCastBear(t, g, me, "GU")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the Bear has %d counters, want 3 (two, plus one from Hardened Scales on the entry)", got)
	}
}
