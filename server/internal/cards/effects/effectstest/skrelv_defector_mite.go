package effectstest

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ToxicOn(g *game.Game, id uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked(); c, _ := g.LookupCardForEffect(id); n = game.ToxicTotal(&c) })
	return n
}
