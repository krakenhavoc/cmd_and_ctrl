package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gptB1Seats is the active seat and the next one.
func GptB1Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// gptB1DamagePlayer deals `n` noncombat damage to `player` from a
// creature of their own.
func GptB1DamagePlayer(g *game.Game, player uuid.UUID, n int) {
	src := B12Creature(g, player, "Pinger", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, player, n) })
}
