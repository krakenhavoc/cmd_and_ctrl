package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heed the Mists — Sorcery — Arcane {3}{U}{U} (Betrayers of Kamigawa, #5602):
//
//	"Mill a card, then draw cards equal to the milled card's mana value."
//
// The mill is Seek Thrills' (MillToZone with a keyed tail), and the draw
// reads the mana value of the card that was milled, as it is in the graveyard
// (a card with {X} in its cost counts X as 0, CR 202.3e; a land is 0). With an
// empty library nothing is milled and nothing is drawn. The draw is the
// spell's own instruction, so a draw that finds the library short loses the
// game as for any draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49330aaf-c005-4e07-ae3a-2579bea3c326",
		Name:         "Heed the Mists",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{N: 1, ThenTail: ctx.ItemTail(heedTheMistsMilled)}.Apply(ctx)
		},
	})
}

// heedTheMistsMilled draws cards equal to the mana value of the card milled
// this way. An on-disk key.
var heedTheMistsMilled = effects.CardTail("s82-old-b600/heed-the-mists-after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			n += c.ManaValue()
		}
	}
	return effects.DrawCards{N: n}.Apply(ctx)
})
