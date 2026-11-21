package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Scavenging Ooze ----------------------------------------------

// oozeEats puts a Scavenging Ooze on the battlefield under `me` and
// activates its ability at `prey`, resolving the ability.
func OozeEats(t *testing.T, g *game.Game, me *game.Player, ooze, prey uuid.UUID) {
	t.Helper()
	B10AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: prey}},
	}); err != nil {
		t.Fatalf("activate Scavenging Ooze: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

func PushOoze(g *game.Game, me *game.Player) uuid.UUID {
	return B12Push(g, me.ID, "Scavenging Ooze", "Creature — Ooze", B14ScavengingOozeOracle, 2, 2)
}

// --- Cling to Dust -------------------------------------------------

const ClingToDustOracle = "9c44e556-4c7a-48bf-a108-c584f69c2cfa"
