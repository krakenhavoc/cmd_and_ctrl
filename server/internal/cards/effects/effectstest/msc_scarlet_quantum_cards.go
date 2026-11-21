package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MscSqInExile(g *game.Game, id uuid.UUID) bool {
	var in bool
	g.ReadSnapshot(func() { in = g.Exile.Contains(id) })
	return in
}
