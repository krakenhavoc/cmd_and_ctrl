package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const sporeFrogOracle = "97db6c39-e690-49b6-93a6-e51b8dfad10b"

// TestSporeFrogSacrificesToFogCombatDamage — the sacrifice pays, the
// Frog leaves, and the rest of the turn's combat damage is prevented.
func TestSporeFrogSacrificesToFogCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frog := pushCatalogPermanent(g, me.ID, "Spore Frog", "Creature — Frog", sporeFrogOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, frog, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(frog) {
		t.Error("Spore Frog should be sacrificed")
	}
	if scopedReplacementCount(g) != 1 {
		t.Fatalf("scoped replacements = %d, want 1 (the Fog effect)", scopedReplacementCount(g))
	}

	defenderID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: defenderID, Name: "Defender", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	attackerID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: attackerID, Name: "Attacker", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	if err := g.MarkCombatDamage(attackerID, defenderID, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}

	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == defenderID {
				marked = c.DamageMarked
			}
		}
	})
	if marked != 0 {
		t.Errorf("DamageMarked = %d, want 0 (Spore Frog should have prevented it)", marked)
	}
}
