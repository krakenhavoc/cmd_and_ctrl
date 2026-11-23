package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A shield bigger than the damage absorbs all of it and keeps the
// remainder for the next event (CR 615.8).
func TestMendingHandsAbsorbsAcrossTwoEvents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := effectstest.PushBear(g, me, "Shielded Bear", 10)

	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3)
	})
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Fatalf("3 into a 4-shield: marked = %d, want 0", got)
	}
	// 1 charge left: a second 3-damage event gets 1 prevented and 2
	// through.
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3)
	})
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("second hit against the 1 remaining charge: marked = %d, want 2", got)
	}
	// Shield exhausted — everything lands now.
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3)
	})
	if got := effectstest.DamageMarkedOn(g, bear); got != 5 {
		t.Errorf("spent shield must not prevent anything: marked = %d, want 5", got)
	}
}

// The partial case, in one event: 6 damage into a 4-point shield is
// 2 damage marked, not 0 and not 6.
func TestMendingHandsPreventsOnlyItsCharge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	bear := effectstest.PushBear(g, me, "Shielded Bear", 10)

	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 6)
	})
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("6 into a 4-shield: marked = %d, want 2", got)
	}
}

// The shield names one object. Damage to anything else is untouched
// — this is what makes it a shield rather than a Fog.
func TestMendingHandsShieldsOnlyItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	shielded := effectstest.PushBear(g, me, "Shielded Bear", 10)
	bystander := effectstest.PushBear(g, me, "Bystander Bear", 10)

	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: shielded}})
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bystander, 3)
	})
	if got := effectstest.DamageMarkedOn(g, bystander); got != 3 {
		t.Errorf("an untargeted creature must take its damage: marked = %d, want 3", got)
	}
	if got := effectstest.DamageMarkedOn(g, shielded); got != 0 {
		t.Errorf("shielded creature took %d damage from an event aimed elsewhere", got)
	}
}

// "Any target" includes players, and the shield serves them with no
// special case: DamageTarget is one UUID whether it names a card or
// a seat.
func TestMendingHandsOnAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID

	effectstest.CastCatalogSpell(t, g, "Mending Hands", "Instant", effectstest.MendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me}})
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(uuid.New(), me, 6)
	})
	if got := effectstest.LifeTotalOf(g, me); got != 38 {
		t.Errorf("6 into a 4-shield on a player: life = %d, want 38", got)
	}
}

// TestTangleDeclaresItsUntapClause pins the #751 closeout.
func TestTangleDeclaresItsUntapClause(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.TangleOracle)
	if !ok {
		t.Fatal("Tangle is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %v, want CompletenessFull", spec.Completeness)
	}
	if len(spec.Caveats) != 0 {
		t.Errorf("Tangle caveats = %v, want none", spec.Caveats)
	}
}

// Holy Day and Tangle are Fog in two other colours and share its
// primitive; the assertion that earns its keep is that they
// register the same uncharged shield, which is what stops a future
// refactor from quietly giving the shared primitive a charge.
func TestHolyDayAndTangleAreFogShaped(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Holy Day", effectstest.HolyDayOracle},
		{"Tangle", effectstest.TangleOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0].ID
			defender := effectstest.PushBear(g, me, "Defender", 2)
			attacker := effectstest.PushBear(g, g.Seats[1].ID, "Attacker", 2)

			effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)

			if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
				t.Fatalf("MarkCombatDamage: %v", err)
			}
			if got := effectstest.DamageMarkedOn(g, defender); got != 0 {
				t.Errorf("combat damage marked = %d, want 0", got)
			}
			// Uncharged: a second combat-damage event this turn is
			// prevented too.
			if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
				t.Fatalf("MarkCombatDamage: %v", err)
			}
			if got := effectstest.DamageMarkedOn(g, defender); got != 0 {
				t.Errorf("a fog has no charges: marked = %d, want 0", got)
			}
			// Non-combat damage still lands.
			if err := g.MarkDamage(defender, 1); err != nil {
				t.Fatalf("MarkDamage: %v", err)
			}
			if got := effectstest.DamageMarkedOn(g, defender); got != 1 {
				t.Errorf("non-combat damage marked = %d, want 1", got)
			}
		})
	}
}
