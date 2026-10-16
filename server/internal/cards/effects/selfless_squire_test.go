package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// selfless_squire_test.go — CR 615.13 (#2047): Selfless Squire's
// "Whenever damage that would be dealt to you is prevented, put that many
// +1/+1 counters on this creature."

const selflessSquireOracle = "7d079c78-b2a9-444f-ab54-c61c041baa6d"

// squireTriggers counts the Squire's triggered items waiting or on the
// stack.
func squireTriggers(g *game.Game, squire uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.SourceCardID == squire {
				n++
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered && it.SourceCardID == squire {
				n++
			}
		}
	})
	return n
}

// castSquire casts the Squire for seat 0 and resolves it and its enters
// trigger: no damage to that player this turn.
func castSquire(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Selfless Squire", "Creature — Human Soldier", selflessSquireOracle, nil)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, id) {
		t.Fatal("the Squire did not resolve")
	}
	return id
}

// Two sources' damage in one instance, stopped by the Squire's own
// shield, is one application: one trigger, for all of it (CR 615.13).
// A second instruction is a second trigger.
func TestSelflessSquireOneTriggerPerApplication(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, opp.ID, "A", 3, "R")
	b := pr7Creature(g, opp.ID, "B", 2, "R")
	squire := castSquire(t, g)
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
	if n := squireTriggers(g, squire); n != 1 {
		t.Fatalf("%d triggers, want one for the one application", n)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d, want %d: the shield prevents it all", me.Life, life)
	}
	if n := countersOn(g, squire, game.CounterPlusOne); n != 5 {
		t.Fatalf("%d counters, want 5", n)
	}
	pr7Hit(t, g, a, me.ID, 4)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, squire, game.CounterPlusOne); n != 9 {
		t.Fatalf("%d counters, want 9 after a second instance of 4", n)
	}
}

// Damage that can't be prevented is dealt, and the Squire gets nothing
// (CR 615.12: the shield prevents nothing).
func TestSelflessSquireNothingForUnpreventableDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, opp.ID, "A", 3, "R")
	squire := castSquire(t, g)
	life := me.Life
	g.WithWriteLock(func() { g.DamageCantBePreventedThisTurnForEffect(uuid.Nil, "Skullcrack") })
	pr7Hit(t, g, a, me.ID, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Fatalf("life %d, want %d", me.Life, life-3)
	}
	if n := countersOn(g, squire, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters, want none", n)
	}
	if n := squireTriggers(g, squire); n != 0 {
		t.Fatalf("%d triggers, want none", n)
	}
}

// Any prevention effect counts, not only the Squire's own: Riot Control's
// shield grows a Squire that was already on the battlefield. Damage
// prevented to another player, or to a creature you control, does not.
func TestSelflessSquireCountsOtherPreventionAndOnlyYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, opp.ID, "A", 3, "R")
	squire := pushCatalogPermanent(g, me.ID, "Selfless Squire", "Creature — Human Soldier", selflessSquireOracle, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	g.WithWriteLock(func() {
		g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			Controller: opp.ID, ProtectPlayer: opp.ID, Label: "their shield",
		})
		g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			Controller: me.ID, ProtectPermanent: bear, Label: "the bear's shield",
		})
	})
	pr7Hit(t, g, a, opp.ID, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(a, bear, 2) })
	g.RunStateChecksForTest()
	if n := squireTriggers(g, squire); n != 0 {
		t.Fatalf("%d triggers, want none: neither damage would have been dealt to you", n)
	}
	castCatalogSpell(t, g, "Riot Control", "Instant", pr7aRiotControl, nil)
	passPriorityAroundTable(t, g)
	pr7Hit(t, g, a, me.ID, 3)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, squire, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters, want 3 from Riot Control's prevention", n)
	}
}
