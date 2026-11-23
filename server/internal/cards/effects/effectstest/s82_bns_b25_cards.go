package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bnsB25Heroic casts Hearts on Fire at the creature and lets it resolve.
func BnsB25Heroic(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, g, "Hearts on Fire", "Instant", J25b1HeartsOnFireOracle, CardRefs(target))
}
