package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const skysovereignConsulFlagshipOracle = "50b14338-9318-4327-a1dd-c0ef38903cc4"

// TestSkysovereignDealsDamageOnETB pins the "enters or attacks" shape
// (AGENTS.md §7, the Sun Titan pattern): entering puts the ability on
// the stack, the controller picks the target as it goes on the stack
// (CR 603.3d), and 3 damage lands on resolution. The victim is a 6/6
// so the 3 damage does not kill it — a dead permanent's marked damage
// is cleared on the way to the graveyard (CR 400.7), which would hide
// the very thing this test checks.
func TestSkysovereignDealsDamageOnETB(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Bear", TypeLine: "Creature — Bear",
		Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Skysovereign, Consul Flagship", "Legendary Artifact — Vehicle",
		skysovereignConsulFlagshipOracle, nil)
	passPriorityAroundTable(t, g)

	pickCard(t, g, caster.ID, victim)
	passPriorityAroundTable(t, g)

	if c := b12Card(t, g, victim); c.DamageMarked != 3 {
		t.Errorf("Skysovereign's ETB trigger: DamageMarked = %d, want 3", c.DamageMarked)
	}
}
