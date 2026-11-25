package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// scoped_combat_tax_test.go — #6191: attack and block taxes scoped to the
// taxed creature (Myr Prototype), priced per blocked attacker (Hipparion,
// Awesome Presence), or paid in life (Heat Wave). CR 508.1h, CR 509.1d,
// CR 509.1c, CR 119.4.

const (
	myrTestOracle       = "test-6191-myr-prototype"
	hipparionTestOracle = "test-6191-hipparion"
	heatWaveTestOracle  = "test-6191-heat-wave"
)

func perCounterMana(n int) string {
	if n <= 0 {
		return ""
	}
	out := ""
	for i := 0; i < n; i++ {
		out += "{1}"
	}
	return out
}

// TestMyrPrototypeAttacksOnlyAfterPayingPerCounter: two +1/+1 counters,
// so attacking costs {2}; refused with {1}, accepted with {2}.
func TestMyrPrototypeAttacksOnlyAfterPayingPerCounter(t *testing.T) {
	g := newActiveGame(t)
	prev := CatalogAttackTaxes
	t.Cleanup(func() { CatalogAttackTaxes = prev })
	CatalogAttackTaxes = func(key string) []AttackTax {
		if key != myrTestOracle {
			return nil
		}
		return []AttackTax{{
			Label:      "This creature can't attack unless you pay {1} for each +1/+1 counter on it.",
			OnAttacker: true,
			AppliesTo:  func(q AttackTaxQuery) bool { return q.Attacker.InstanceID == q.Source.InstanceID },
			ManaCost:   func(q AttackTaxQuery) string { return perCounterMana(q.Attacker.Counters[CounterPlusOne]) },
		}}
	}
	me := g.Seats[0]
	myr := pushCombatant(t, g, me, "Myr Prototype", 0, 0)
	c := findCard(g, myr)
	c.OracleID = myrTestOracle
	c.Counters = map[string]int{CounterPlusOne: 2}
	g.WithWriteLock(func() { g.layerVersion.Add(1); g.RecomputeLayersIfStaleLocked() })
	advanceIntoStep(t, g, StepDeclareAttackers)

	addMana(me, 1)
	if err := g.DeclareAttacker(myr, g.Seats[1].ID); !errors.Is(err, ErrAttackTaxUnpaid) {
		t.Fatalf("attack with {1} for two counters: err = %v, want ErrAttackTaxUnpaid", err)
	}
	addMana(me, 1)
	if err := g.DeclareAttacker(myr, g.Seats[1].ID); err != nil {
		t.Fatalf("attack with {2} for two counters: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("pool = %d after the attack, want 0", n)
	}
}

// hipparionTable is seat 0 attacking with a 2/2 and a 3/3, and seat 1
// holding a Hipparion ("can't block creatures with power 3 or greater
// unless you pay {1}") and, when heatWave, a Heat Wave on seat 0's side
// ("Nonblue creatures can't block creatures you control unless their
// controller pays 1 life for each blocking creature they control").
func hipparionTable(t *testing.T, heatWave bool) (g *Game, small, big, hip, wall uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	withCatalogBlockTaxes(t, func(key string) []BlockTax {
		switch key {
		case hipparionTestOracle:
			return []BlockTax{{
				Label:       "This creature can't block creatures with power 3 or greater unless you pay {1}.",
				PerAttacker: true,
				AppliesTo: func(q BlockTaxQuery) bool {
					return BlockedBySelf(q) && q.Attacker.Effective().Power >= 3
				},
				ManaCost: func(BlockTaxQuery) string { return "{1}" },
			}}
		case heatWaveTestOracle:
			return []BlockTax{{
				Label: "Nonblue creatures can't block creatures you control unless their controller pays 1 life for each blocking creature they control.",
				AppliesTo: func(q BlockTaxQuery) bool {
					if q.Blocker.HasColor("U") {
						return false
					}
					for _, a := range q.Blocking {
						if a.Controller == q.Source.Controller {
							return true
						}
					}
					return false
				},
				LifeCost: func(BlockTaxQuery) int { return 1 },
			}}
		}
		return nil
	})
	small = pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	big = pushCombatant(t, g, g.Seats[0], "Ogre", 3, 3)
	hip = pushCombatant(t, g, g.Seats[1], "Hipparion", 1, 3)
	findCard(g, hip).OracleID = hipparionTestOracle
	wall = pushCombatant(t, g, g.Seats[1], "Wall", 0, 4)
	if heatWave {
		hw := NewCard("Heat Wave", g.Seats[0].ID)
		hw.TypeLine = "Enchantment"
		hw.OracleID = heatWaveTestOracle
		g.Battlefield.PushTop(hw)
	}
	declareAttacks(t, g, small, big)
	return g, small, big, hip, wall
}

// TestHipparionBlocksASmallAttackerFreeAndPaysForABigOne is the
// acceptance: Hipparion blocks a 2-power attacker for nothing and pays
// {1} to block a 3-power one.
func TestHipparionBlocksASmallAttackerFreeAndPaysForABigOne(t *testing.T) {
	g, small, big, hip, _ := hipparionTable(t, false)
	def := g.Seats[1]
	var add BlockTaxPrice
	g.ReadSnapshot(func() {
		add = g.PriceBlockAdditionForEffect(def.ID, []BlockDeclaration{{Blocker: hip, Attacker: small}})
	})
	if !add.IsFree() {
		t.Errorf("blocking the 2/2 adds %+v, want nothing", add)
	}
	g.ReadSnapshot(func() { add = g.PriceBlockAdditionForEffect(def.ID, []BlockDeclaration{{Blocker: hip, Attacker: big}}) })
	if add.Cost != "{1}" || len(add.Lines) != 1 || add.Lines[0].Attacker != big {
		t.Errorf("blocking the 3/3 adds %+v, want {1} on the pair", add)
	}

	if err := block(t, g, hip, small); err != nil {
		t.Fatalf("block the 2/2: %v", err)
	}
	if p := g.PriceBlockDeclaration(def.ID); !p.IsFree() {
		t.Fatalf("Hipparion blocking a 2/2 costs %q, want nothing", p.Cost)
	}
	if err := block(t, g, hip, big); err != nil {
		t.Fatalf("re-stage onto the 3/3: %v", err)
	}
	if p := g.PriceBlockDeclaration(def.ID); p.Cost != "{1}" {
		t.Fatalf("Hipparion blocking a 3/3 costs %q, want {1}", p.Cost)
	}
	blockTaxErr(t, g.FinishBlocks(def.ID), "{1}")
	addMana(def, 1)
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("finish with {1}: %v", err)
	}
	if findCard(g, hip).BlockingTarget != big {
		t.Error("the paid block is not standing")
	}
}

// TestHeatWaveChargesALifePerNonblueBlocker: each nonblue creature
// blocking a creature Heat Wave's controller controls costs 1 life, paid
// as the declaration completes; with too little life it is refused
// (CR 119.4) and nothing is paid.
func TestHeatWaveChargesALifePerNonblueBlocker(t *testing.T) {
	g, small, big, hip, wall := hipparionTable(t, true)
	def := g.Seats[1]
	if err := block(t, g, hip, small, wall, big); err != nil {
		t.Fatalf("block: %v", err)
	}
	p := g.PriceBlockDeclaration(def.ID)
	// Hipparion blocks the 2/2, which it may do for free; the Wall blocks
	// the 3/3, which no tax of its own prices.
	if p.Life != 2 || p.Cost != "" {
		t.Fatalf("price = %+v, want 2 life (two nonblue blockers) and no mana", p)
	}
	g.WithWriteLock(func() { def.Life = 1 })
	addMana(def, 1)
	if err := g.FinishBlocks(def.ID); !errors.Is(err, ErrBlockTaxUnpaid) {
		t.Fatalf("finish at 1 life owing 2: err = %v, want ErrBlockTaxUnpaid", err)
	}
	if def.Life != 1 || def.ManaPool.Total() != 1 {
		t.Errorf("a refused finish paid something: life %d, pool %d", def.Life, def.ManaPool.Total())
	}
	g.WithWriteLock(func() { def.Life = 20 })
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatalf("finish at 20 life: %v", err)
	}
	if def.Life != 18 {
		t.Errorf("life = %d, want 18 (1 per nonblue blocker)", def.Life)
	}
}

// TestPerAttackerBlockTaxIsNotOwedByARequirement: CR 509.1c — a pair a
// block tax would charge for is not required. (Checked through the
// engine's own predicate.)
func TestPerAttackerBlockTaxIsNotOwedByARequirement(t *testing.T) {
	g, small, big, hip, _ := hipparionTable(t, false)
	g.ReadSnapshot(func() {
		if g.blockPairTaxedLocked(findBattlefieldCard(g, hip), findBattlefieldCard(g, small)) {
			t.Error("Hipparion blocking a 2/2 is taxed")
		}
		if !g.blockPairTaxedLocked(findBattlefieldCard(g, hip), findBattlefieldCard(g, big)) {
			t.Error("Hipparion blocking a 3/3 is not taxed")
		}
	})
}
