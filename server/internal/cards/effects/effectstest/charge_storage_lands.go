package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PoolColours(p *game.Player) map[string]int {
	out := map[string]int{}
	for _, tok := range p.ManaPool.Units() {
		out[tok.Color]++
	}
	return out
}

func AddCountersForTest(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}
