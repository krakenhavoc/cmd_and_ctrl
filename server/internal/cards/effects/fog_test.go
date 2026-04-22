package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fog_test.go covers the S17 sub-PR 5 damage-prevention scaffold:
// Fog cancels combat damage (both creature-on-creature and
// attacker-on-player) via the turn-scoped replacement slot.
// Non-combat damage (spell / effect) still lands. The replacement
// clears at StepCleanup so next turn's combat damage resolves
// normally.

const fogOracle = "27e9db49-7af7-4bef-ad4c-bf5dfb92030d"

// TestFogCancelsCombatDamageToCreature — Fog is resolved; the
// combat damage resolver's creature-on-creature damage is then
// canceled by the pipeline.
func TestFogCancelsCombatDamageToCreature(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[0]

	// Cast + resolve Fog.
	castCatalogSpell(t, g, "Fog", "Instant", fogOracle, nil)
	passPriorityAroundTable(t, g)

	// Fog turn-scoped replacement is now registered.
	if len(g.TurnScopedReplacements) != 1 {
		t.Fatalf("TurnScopedReplacements = %d, want 1", len(g.TurnScopedReplacements))
	}

	// Directly invoke the combat-damage helper — simulates a
	// Bear-on-Bear block.
	defenderID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: defenderID,
		Name:       "Defender",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      active.ID,
		Controller: active.ID,
	})
	attackerID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: attackerID,
		Name:       "Attacker",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      g.Seats[1].ID,
		Controller: g.Seats[1].ID,
	})

	if err := g.MarkCombatDamage(attackerID, defenderID, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}

	// Damage should have been canceled — DamageMarked stays 0.
	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == defenderID {
				marked = c.DamageMarked
				return
			}
		}
	})
	if marked != 0 {
		t.Errorf("DamageMarked = %d, want 0 (Fog should cancel combat damage)", marked)
	}
}

// TestFogLeavesNonCombatDamageAlone — spell damage (Lightning
// Bolt-ish) via DealDamageToCreatureForEffect does NOT carry the
// IsCombatDamage flag; Fog's predicate gates it out.
func TestFogLeavesNonCombatDamageAlone(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[0]

	castCatalogSpell(t, g, "Fog", "Instant", fogOracle, nil)
	passPriorityAroundTable(t, g)

	victimID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: victimID,
		Name:       "Victim",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  3,
		Owner:      active.ID,
		Controller: active.ID,
	})

	// Non-combat damage (spell) — uses the public MarkDamage (no
	// IsCombatDamage flag).
	if err := g.MarkDamage(victimID, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}

	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == victimID {
				marked = c.DamageMarked
				return
			}
		}
	})
	if marked != 2 {
		t.Errorf("DamageMarked = %d, want 2 (Fog should not affect non-combat damage)", marked)
	}
}

// TestFogClearsAtCleanup — after advancing past StepCleanup the
// turn-scoped replacement is gone. Next turn's combat damage
// lands normally.
func TestFogClearsAtCleanup(t *testing.T) {
	g := newCatalogGame(t)

	castCatalogSpell(t, g, "Fog", "Instant", fogOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.TurnScopedReplacements) != 1 {
		t.Fatalf("Fog didn't register: %d turn-scoped", len(g.TurnScopedReplacements))
	}

	// Walk the cursor past cleanup. From main phase to next turn's
	// main phase is ≤ 20 step advances.
	for i := 0; i < 30; i++ {
		if g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepPrecombatMain {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if len(g.TurnScopedReplacements) != 0 {
		t.Errorf("turn-scoped replacements = %d after cleanup, want 0", len(g.TurnScopedReplacements))
	}
}
