package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Two sources' damage in one instance, stopped by the Squire's own
// shield, is one application: one trigger, for all of it (CR 615.13).
// A second instruction is a second trigger.
func TestSelflessSquireOneTriggerPerApplication(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, opp.ID, "A", 3, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 2, "R")
	squire := effectstest.CastSquire(t, g)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToPlayerForEffect(a, me.ID, 3); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(b, me.ID, 2)
		})
	})
	g.RunStateChecksForTest()
	if n := effectstest.SquireTriggers(g, squire); n != 1 {
		t.Fatalf("%d triggers, want one for the one application", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d, want %d: the shield prevents it all", me.Life, life)
	}
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 5 {
		t.Fatalf("%d counters, want 5", n)
	}
	effectstest.Pr7Hit(t, g, a, me.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 9 {
		t.Fatalf("%d counters, want 9 after a second instance of 4", n)
	}
}

// Damage that can't be prevented is dealt, and the Squire gets nothing
// (CR 615.12: the shield prevents nothing).
func TestSelflessSquireNothingForUnpreventableDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, opp.ID, "A", 3, "R")
	squire := effectstest.CastSquire(t, g)
	life := me.Life
	g.WithWriteLock(func() { g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack") })
	effectstest.Pr7Hit(t, g, a, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d", me.Life, life-3)
	}
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters, want none", n)
	}
	if n := effectstest.SquireTriggers(g, squire); n != 0 {
		t.Fatalf("%d triggers, want none", n)
	}
}

// Spider-Punk's "Damage can't be prevented" switches the trigger off with
// the prevention: the damage is dealt and nothing triggers.
func TestSelflessSquireNothingUnderSpiderPunk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, opp.ID, "A", 3, "R")
	squire := effectstest.CastSquire(t, g)
	effectstest.PushCatalogPermanent(g, opp.ID, "Spider-Punk", "Legendary Creature — Spider Human Hero", "03bf7a90-78c1-4e83-8553-e22d85e49876", false)
	life := me.Life
	effectstest.Pr7Hit(t, g, a, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d", me.Life, life-3)
	}
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 0 || effectstest.SquireTriggers(g, squire) != 0 {
		t.Fatalf("%d counters, %d triggers, want none", n, effectstest.SquireTriggers(g, squire))
	}
}

// Any prevention effect counts, not only the Squire's own: Riot Control's
// shield grows a Squire that was already on the battlefield. Damage
// prevented to another player, or to a creature you control, does not.
func TestSelflessSquireCountsOtherPreventionAndOnlyYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, opp.ID, "A", 3, "R")
	squire := effectstest.PushCatalogPermanent(g, me.ID, "Selfless Squire", "Creature — Human Soldier", effectstest.SelflessSquireOracle, false)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	g.WithWriteLock(func() {
		g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			Controller: opp.ID, ProtectPlayer: opp.ID, Label: "their shield",
		})
		g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			Controller: me.ID, ProtectPermanent: bear, Label: "the bear's shield",
		})
	})
	effectstest.Pr7Hit(t, g, a, opp.ID, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(a, bear, 2) })
	g.RunStateChecksForTest()
	if n := effectstest.SquireTriggers(g, squire); n != 0 {
		t.Fatalf("%d triggers, want none: neither damage would have been dealt to you", n)
	}
	effectstest.CastCatalogSpell(t, g, "Riot Control", "Instant", effectstest.Pr7aRiotControl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr7Hit(t, g, a, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters, want 3 from Riot Control's prevention", n)
	}
}
