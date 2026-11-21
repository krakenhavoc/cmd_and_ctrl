package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nameless_inversion_test.go — Nameless Inversion (#4691, Lorwyn
// Eclipsed #4458).

const NamelessInversionOracle = "d6911456-42cb-4b23-a560-dc98784530f5"

func CastNamelessInversion(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, g, "Nameless Inversion", "Kindred Instant — Shapeshifter",
		NamelessInversionOracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	PassPriorityAroundTable(t, g)
}
