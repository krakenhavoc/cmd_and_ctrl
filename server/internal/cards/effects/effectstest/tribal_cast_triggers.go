package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func CastFree(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	CastFromHandForTest(t, g, p, name, typeLine, "", "", game.CastSpellParams{FromZone: "hand"})
	PassPriorityAroundTable(t, g)
}
