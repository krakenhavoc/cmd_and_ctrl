package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const magmaticGalleonOracle = "59322432-591a-4e8d-aff5-12ca1feb1028"

// TestMagmaticGalleonETBDamageOverkillMakesATreasure exercises both
// abilities off one entry: the ETB deals 5 to a 3-toughness opponent
// creature (2 excess), which is enough noncombat damage to kill it,
// and the second ability's excess watch fires off that same event.
func TestMagmaticGalleonETBDamageOverkillMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Squishy", 2, 3)

	castCatalogSpell(t, g, "Magmatic Galleon", "Artifact — Vehicle", magmaticGalleonOracle, nil)
	passPriorityAroundTable(t, g)

	if latestTriggerPrompt(g, caster.ID) != nil {
		t.Errorf("Magmatic Galleon's ETB damage is mandatory — no yes/no prompt")
	}
	prompt := latestPickTarget(g, caster.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for the ETB damage")
	}
	pickCard(t, g, caster.ID, victim)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("5 damage to a 3-toughness creature should have destroyed it")
	}
	if n := b43TokensNamed(g, caster.ID, "Treasure"); n != 1 {
		t.Fatalf("Treasures for the controller = %d, want 1 (the overkill)", n)
	}
}

// TestMagmaticGalleonNoOverkillNoTreasure — 5 damage to a creature
// with enough toughness to survive it makes no Treasure: nothing was
// dealt in excess of what was needed to destroy it (and nothing was,
// since it didn't die).
func TestMagmaticGalleonNoOverkillNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Big Wall", 0, 10)

	castCatalogSpell(t, g, "Magmatic Galleon", "Artifact — Vehicle", magmaticGalleonOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, caster.ID, victim)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Fatal("5 damage to a 10-toughness creature should not have destroyed it")
	}
	if n := b43TokensNamed(g, caster.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0 (no excess damage)", n)
	}
}

// TestMagmaticGalleonCombatDamageIsNotExcessWatch — the second
// ability is explicitly NONCOMBAT damage; a combat kill, however
// overkilled, must not mint a Treasure.
func TestMagmaticGalleonCombatDamageDoesNotMakeATreasure(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, caster.ID, "Attacker", 8, 8)
	victim := pushVanillaCreature(g, opp.ID, "Squishy", 2, 3)

	castCatalogSpell(t, g, "Magmatic Galleon", "Artifact — Vehicle", magmaticGalleonOracle, nil)
	passPriorityAroundTable(t, g)
	// Decline the ETB's mandatory pick is not possible (no legal
	// alternative here — there's exactly one opponent creature), so
	// resolve it against the same victim first...
	pickCard(t, g, caster.ID, victim)
	passPriorityAroundTable(t, g)

	// ...then deal combat damage from an unrelated attacker to a
	// FRESH creature to isolate the combat-damage case.
	victim2 := pushVanillaCreature(g, opp.ID, "Squishy Two", 2, 3)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:   game.EventDealDamage,
			Source: attacker,
			Target: victim2,
			Amount: 8,
			Combat: true,
		})
	})
	before := b43TokensNamed(g, caster.ID, "Treasure")
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, caster.ID, "Treasure"); got != before {
		t.Errorf("combat damage minted a Treasure: %d -> %d", before, got)
	}
}
