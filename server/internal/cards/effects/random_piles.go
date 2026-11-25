package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// random_piles.go — Whims of the Fates (#6248): "Starting with you, each
// player separates all permanents they control into three piles. Then
// each player chooses one of their piles at random and sacrifices those
// permanents." The engine half is game/random_piles.go.

var randomPilesSacrifice = CardTail("tail/random-piles-sacrifice", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), o.IDs, game.KeyedTail{})
})

// EachPlayerSacrificesARandomPile is the whole of Whims of the Fates'
// text, with `piles` the number of piles and `label` the card's name.
func EachPlayerSacrificesARandomPile(piles int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return g.SeparateIntoPilesPickOneAtRandomForEffect(game.RandomPiles{
			Source:  ctx.Source(),
			Players: playersInTurnOrderFrom(g, ctx.Controller()),
			Piles:   piles,
			Label:   label,
			Then:    ctx.Tail(randomPilesSacrifice),
		})
	}
}

// playersInTurnOrderFrom is every player still in the game in the game's
// current turn order (reversed or not), starting with `first`.
func playersInTurnOrderFrom(g *game.Game, first uuid.UUID) []uuid.UUID {
	start := 0
	for i, s := range g.Seats {
		if s != nil && s.ID == first {
			start = i
		}
	}
	var out []uuid.UUID
	for i := 0; i < len(g.Seats); i++ {
		p := g.Seats[g.SeatInTurnOrderForEffect(start, i)]
		if p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}
