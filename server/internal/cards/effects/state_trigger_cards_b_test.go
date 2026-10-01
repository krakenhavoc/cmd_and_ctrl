package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_cards_b_test.go — the second batch of ADR 0107 §1's
// state-trigger cards (#1858): the tide pair, the damage-to-counters
// pair, The Millennium Calendar and Garruk Relentless.

const (
	stHomarid        = "5d02b0d9-cc53-4967-84bb-8b7f5c489b5c"
	stTidalInfluence = "b1393388-b8b3-4c54-a086-f5e6d9908972"
	stForceBubble    = "655ae8e7-372b-4d8d-b33f-4aca46831abb"
	stMillenniumCal  = "1f250443-8d5e-46c9-920e-8e9373780e32"
)

func stAddCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

func stDamagePlayer(t *testing.T, g *game.Game, source, player uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, player, n); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
}

// Homarid's tide cycle: 1/1, 2/2, 3/3, and the fourth counter triggers
// the reset once.
func TestHomaridTideCycle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	h := stCard("Homarid", stHomarid, "Creature — Homarid", 2, 2)
	h.Counters = map[string]int{"tide": 1} // as the entry replacement leaves it
	id := apaPush(g, me.ID, me.ID, h)
	pt := func() (int, int) { c := apaLive(g, id); return c.CurrentPower(), c.CurrentToughness() }
	if p, tt := pt(); p != 1 || tt != 1 {
		t.Fatalf("one tide counter: %d/%d, want 1/1", p, tt)
	}
	stAddCounters(t, g, id, "tide", 1)
	if p, tt := pt(); p != 2 || tt != 2 {
		t.Fatalf("two tide counters: %d/%d, want 2/2", p, tt)
	}
	stAddCounters(t, g, id, "tide", 1)
	if p, tt := pt(); p != 3 || tt != 3 {
		t.Fatalf("three tide counters: %d/%d, want 3/3", p, tt)
	}
	stAddCounters(t, g, id, "tide", 1)
	if n := stStateItems(g, id, "Homarid — remove"); n != 1 {
		t.Fatalf("the fourth counter queued %d resets, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if c := apaLive(g, id); c.Counters["tide"] != 0 {
		t.Fatalf("after the reset: %d tide counters, want 0", c.Counters["tide"])
	}
}

// Homarid enters with its tide counter (CR 614.1c).
func TestHomaridEntersWithATideCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Homarid", TypeLine: "Creature — Homarid", OracleID: stHomarid,
		Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := apaLive(g, id); c == nil || c.Counters["tide"] != 1 {
		t.Fatalf("Homarid entered with %+v, want one tide counter", c)
	}
}

// Tidal Influence pumps and shrinks every blue creature by its tide
// count, and a second one can't be cast while it is out.
func TestTidalInfluenceAffectsBlueCreaturesAndBlocksASecondCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	ti := castCatalogSpell(t, g, "Tidal Influence", "Enchantment", stTidalInfluence, nil)
	passPriorityAroundTable(t, g)
	blue := stCard("Merfolk", "", "Creature — Merfolk", 2, 2)
	blue.Colors = []string{"U"}
	b := apaPush(g, bob.ID, bob.ID, blue)
	red := apaPush(g, me.ID, me.ID, stCard("Goblin", "", "Creature — Goblin", 2, 2))
	if c := apaLive(g, b); c.CurrentPower() != 0 {
		t.Fatalf("blue creature power with one tide counter = %d, want 0", c.CurrentPower())
	}
	if c := apaLive(g, red); c.CurrentPower() != 2 {
		t.Fatalf("a red creature changed: %d", c.CurrentPower())
	}
	stAddCounters(t, g, ti, "tide", 2)
	if c := apaLive(g, b); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Fatalf("blue creature with three tide counters = %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
	if err := castCatalogSpellErr(t, g, "Tidal Influence", "Enchantment", stTidalInfluence, nil); err == nil {
		t.Fatal("a second Tidal Influence was cast while one is on the battlefield")
	}
	stAddCounters(t, g, ti, "tide", 1)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, ti); c.Counters["tide"] != 0 {
		t.Fatalf("after the reset: %d tide counters, want 0", c.Counters["tide"])
	}
}

// Force Bubble turns damage to its controller into depletion counters
// and is sacrificed at four.
func TestForceBubbleAbsorbsDamageUntilFour(t *testing.T) {
	g := newCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	fb := apaPush(g, me.ID, me.ID, stCard("Force Bubble", stForceBubble, "Enchantment", 0, 0))
	src := apaPush(g, bob.ID, bob.ID, stCard("Shock Source", "", "Creature — Goblin", 1, 1))
	life := me.Life
	stDamagePlayer(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d -> %d: the damage was not replaced", life, me.Life)
	}
	if c := apaLive(g, fb); c.Counters["depletion"] != 3 {
		t.Fatalf("depletion counters = %d, want 3", c.Counters["depletion"])
	}
	if n := stStateItems(g, fb, "Force Bubble — sacrifice"); n != 0 {
		t.Fatalf("three counters triggered the sacrifice (%d)", n)
	}
	stDamagePlayer(t, g, src, bob.ID, 2) // someone else's damage is untouched
	stDamagePlayer(t, g, src, me.ID, 2)
	if me.Life != life {
		t.Fatalf("the second hit got through: life %d", me.Life)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, fb) {
		t.Fatal("Force Bubble survived five depletion counters")
	}
}

// Force Bubble's counters go at each end step.
func TestForceBubbleEmptiesAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fb := apaPush(g, me.ID, me.ID, stCard("Force Bubble", stForceBubble, "Enchantment", 0, 0))
	stAddCounters(t, g, fb, "depletion", 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, fb); c == nil || c.Counters["depletion"] != 0 {
		t.Fatalf("after the end step: %+v, want no depletion counters", c)
	}
}

// The Millennium Calendar counts the permanents its controller untapped
// in their untap step, doubles on demand, and wins at 1,000.
func TestMillenniumCalendarCountsUntapsAndDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cal := apaPush(g, me.ID, me.ID, stCard("The Millennium Calendar", stMillenniumCal, "Legendary Artifact", 0, 0))
	for i := 0; i < 3; i++ {
		l := stCard("Forest", "", "Basic Land — Forest", 0, 0)
		l.Tapped = true
		apaPush(g, me.ID, me.ID, l)
	}
	// Another player's tapped permanent untaps in their own step, not ours.
	bobLand := stCard("Forest", "", "Basic Land — Forest", 0, 0)
	bobLand.Tapped = true
	apaPush(g, g.Seats[1].ID, g.Seats[1].ID, bobLand)

	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if c := apaLive(g, cal); c.Counters[game.CounterTime] != 3 {
		t.Fatalf("time counters after the untap step = %d, want 3", c.Counters[game.CounterTime])
	}
	advanceToMain(t, g)
	apaMana(me, "C", "C")
	apaActivate(t, g, me, cal, 0, game.ActivateAbilityParams{})
	if c := apaLive(g, cal); c.Counters[game.CounterTime] != 6 {
		t.Fatalf("time counters after doubling = %d, want 6", c.Counters[game.CounterTime])
	}
	stAddCounters(t, g, cal, game.CounterTime, 994)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, cal) {
		t.Fatal("the Calendar is still on the battlefield at 1,000 counters")
	}
	if g.State == game.StateActive || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("state %v outcome %+v: each opponent lost 1,000 life", g.State, g.Outcome)
	}
}
