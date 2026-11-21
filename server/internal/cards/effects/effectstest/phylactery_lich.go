package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phylactery_lich_test.go — #1893, ADR 0170: an as-enters choice of
// another permanent, and the state trigger that watches it.

const phylacteryLichOracle = "520c4d88-34b6-4a26-81ca-65eb8b10461c"

func CastPhylacteryLich(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Phylactery Lich", "Creature — Zombie", phylacteryLichOracle, nil)
	PassPriorityAroundTable(t, g)
	return id
}
