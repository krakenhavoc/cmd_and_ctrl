package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avB22Has reports whether the battlefield permanent has the keyword
// through the layers.
func AvB22Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := AvB9Live(t, g, id)
	return game.HasKeyword(&c, kw)
}
