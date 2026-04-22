package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// anthem_combat_test.go covers the S16 sub-PR 3 combat regression
// the user caught during manual testing: a 2/2 attacker pumped to
// 3/3 by Glorious Anthem must deal 3 damage, not the printed 2.
// And a 2/2 blocker pumped to 3/3 must take 4 damage from a 4/4
// attacker without dying past its effective toughness.
//
// These exercise the wiring of CurrentPower / CurrentToughness
// through the layer engine's Effective() rather than printed P/T,
// plus the recompute call at the top of ResolveCombatDamage and
// the SBA loop.

// TestAnthemPumpedAttackerDealsBuffedDamage proves the combat-damage
// path now reads post-layer effective power. 2/2 Bears + Glorious
// Anthem = effective 3/3, attacks defending player → defender takes
// 3 damage, not the printed 2.
func TestAnthemPumpedAttackerDealsBuffedDamage(t *testing.T) {
	g := newCatalogGame(t)
	attacker := g.Seats[0]
	defender := g.Seats[1]

	bears := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      attacker.ID,
		Controller: attacker.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   gloriousAnthemOracle,
		Owner:      attacker.ID,
		Controller: attacker.ID,
	})

	// Advance to declare-attackers step, declare the attacker, then
	// advance into combat-damage and let ResolveCombatDamage fire.
	for g.Turn.Step != game.StepDeclareAttackers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare-attackers: %v", err)
		}
	}
	if err := g.DeclareAttacker(bears, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	beforeLife := defender.Life
	// AdvanceStep auto-invokes ResolveCombatDamage on entering the
	// combat-damage step — no manual call needed.
	for g.Turn.Step != game.StepCombatDamage {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to combat-damage: %v", err)
		}
	}

	loss := beforeLife - defender.Life
	if loss != 3 {
		t.Errorf("defender took %d damage from pumped Bears, want 3 (printed 2 + anthem 1)", loss)
	}
}

// TestAnthemPumpedBlockerSurvivesPrintedToughnessLethal proves the
// lethal-damage SBA reads post-layer effective toughness. A 2/2
// creature pumped to 3/3 by an anthem with 2 marked damage stays
// alive: 2 < effective toughness 3. Without the fix the SBA would
// read printed toughness 2 (== marked damage) and trigger the
// lethal SBA, destroying the creature.
func TestAnthemPumpedBlockerSurvivesPrintedToughnessLethal(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0]

	bears := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   gloriousAnthemOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	// MarkDamage runs the SBA loop after the mutation. With effective
	// toughness 3 and 2 marked damage, the lethal SBA must NOT fire.
	if err := g.MarkDamage(bears, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}

	stillThere := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == bears {
				stillThere = true
				return
			}
		}
	})
	if !stillThere {
		t.Errorf("anthem-pumped Bears died to 2 marked damage; should survive (effective toughness 3 > 2)")
	}
}
