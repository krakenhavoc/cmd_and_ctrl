package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_triggers_test.go — S19 sub-PR 7: "deals combat damage to a
// player" triggers. Seat 0 is the active player; its creatures
// attack seat 1. Combat damage lands in the combat damage step
// (AdvanceStep into it resolves damage + runs SBAs + drains
// triggers), so the trigger is on the stack — or its prompt queued
// — as soon as the step is entered.

const (
	BidentOracle      = "e1afaef7-9fa3-4662-a95f-adfb0da9fd11"
	CoastalPiracyOrcl = "8a05ec32-7b0c-4f23-a4f7-413301c2a70a"
	ScrollThiefOracle = "637c5583-4683-4ae4-8b4e-f5da42a772c7"
	EdricOracle       = "9a1de7e4-9930-4db3-a8f3-d146d0abf38b"
	// Lightning Bolt, for the non-combat-damage control case. Named
	// distinctly from the sub-PR 6 test file's constant so the two
	// branches merge without a collision.
	BoltOracleCombat = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// advanceTo walks AdvanceStep until the cursor is on step.
func AdvanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 40 && g.Turn.Step != step; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s: %v", step, err)
		}
	}
	if g.Turn.Step != step {
		t.Fatalf("never reached %s (at %s)", step, g.Turn.Step)
	}
}

// attackWith declares every given creature as attacking defender
// and advances into the combat damage step.
func AttackWith(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", id, err)
		}
	}
	AdvanceTo(t, g, game.StepCombatDamage)
}

// pushVanillaCreature seeds a non-catalog creature that can attack
// (EnteredBattlefieldAt is zero, so no summoning sickness).
func PushVanillaCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return PushDiesCreatureForTest(g, owner, name, "", "Creature — Test", power, toughness)
}
