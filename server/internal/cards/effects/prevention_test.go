package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prevention_test.go — CR 615.8, the charged shield. fog_test.go
// already pins the uncharged Fog shape; what is new in S30 is the
// arithmetic, and the arithmetic has exactly three interesting
// cases: the shield covers the damage, the shield is smaller than
// the damage, and the shield has already been spent.
//
// The middle one is the case the generous first draft gets wrong. A
// 4-point shield facing 6 damage must let 2 through; an
// implementation that cancels the event whenever the shield covers
// any of it prevents all 6 and passes every test that only checks
// "did the creature survive".

const (
	holyDayOracle     = "98423a34-f044-4811-b288-56981d604b6e"
	tangleOracle      = "f627e125-15af-4e53-b34e-82b60e4ec87b"
	mendingHandsOracl = "a612f30d-cd55-438b-a7de-8c80509183aa"
)

func pushBear(g *game.Game, owner uuid.UUID, name string, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}

func lifeTotalOf(g *game.Game, playerID uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == playerID {
			return p.Life
		}
	}
	return -1
}

func damageMarkedOn(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

// A shield bigger than the damage absorbs all of it and keeps the
// remainder for the next event (CR 615.8).
func TestMendingHandsAbsorbsAcrossTwoEvents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	bear := pushBear(g, me, "Shielded Bear", 10)

	castCatalogSpell(t, g, "Mending Hands", "Instant", mendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3)
	})
	if got := damageMarkedOn(g, bear); got != 0 {
		t.Fatalf("3 into a 4-shield: marked = %d, want 0", got)
	}
	// 1 charge left: a second 3-damage event gets 1 prevented and 2
	// through.
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3)
	})
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("second hit against the 1 remaining charge: marked = %d, want 2", got)
	}
	// Shield exhausted — everything lands now.
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3)
	})
	if got := damageMarkedOn(g, bear); got != 5 {
		t.Errorf("spent shield must not prevent anything: marked = %d, want 5", got)
	}
}

// The partial case, in one event: 6 damage into a 4-point shield is
// 2 damage marked, not 0 and not 6.
func TestMendingHandsPreventsOnlyItsCharge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	bear := pushBear(g, me, "Shielded Bear", 10)

	castCatalogSpell(t, g, "Mending Hands", "Instant", mendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 6)
	})
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("6 into a 4-shield: marked = %d, want 2", got)
	}
}

// The shield names one object. Damage to anything else is untouched
// — this is what makes it a shield rather than a Fog.
func TestMendingHandsShieldsOnlyItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID
	shielded := pushBear(g, me, "Shielded Bear", 10)
	bystander := pushBear(g, me, "Bystander Bear", 10)

	castCatalogSpell(t, g, "Mending Hands", "Instant", mendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: shielded}})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.New(), bystander, 3)
	})
	if got := damageMarkedOn(g, bystander); got != 3 {
		t.Errorf("an untargeted creature must take its damage: marked = %d, want 3", got)
	}
	if got := damageMarkedOn(g, shielded); got != 0 {
		t.Errorf("shielded creature took %d damage from an event aimed elsewhere", got)
	}
}

// "Any target" includes players, and the shield serves them with no
// special case: DamageTarget is one UUID whether it names a card or
// a seat.
func TestMendingHandsOnAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID

	castCatalogSpell(t, g, "Mending Hands", "Instant", mendingHandsOracl,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me}})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(uuid.New(), me, 6)
	})
	if got := lifeTotalOf(g, me); got != 38 {
		t.Errorf("6 into a 4-shield on a player: life = %d, want 38", got)
	}
}

// Holy Day and Tangle are Fog in two other colours and share its
// primitive; the assertion that earns its keep is that they
// register the same uncharged shield, which is what stops a future
// refactor from quietly giving the shared primitive a charge.
func TestHolyDayAndTangleAreFogShaped(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Holy Day", holyDayOracle},
		{"Tangle", tangleOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0].ID
			defender := pushBear(g, me, "Defender", 2)
			attacker := pushBear(g, g.Seats[1].ID, "Attacker", 2)

			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, nil)
			passPriorityAroundTable(t, g)

			if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
				t.Fatalf("MarkCombatDamage: %v", err)
			}
			if got := damageMarkedOn(g, defender); got != 0 {
				t.Errorf("combat damage marked = %d, want 0", got)
			}
			// Uncharged: a second combat-damage event this turn is
			// prevented too.
			if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
				t.Fatalf("MarkCombatDamage: %v", err)
			}
			if got := damageMarkedOn(g, defender); got != 0 {
				t.Errorf("a fog has no charges: marked = %d, want 0", got)
			}
			// Non-combat damage still lands.
			if err := g.MarkDamage(defender, 1); err != nil {
				t.Fatalf("MarkDamage: %v", err)
			}
			if got := damageMarkedOn(g, defender); got != 1 {
				t.Errorf("non-combat damage marked = %d, want 1", got)
			}
		})
	}
}
