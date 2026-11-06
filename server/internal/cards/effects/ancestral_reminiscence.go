package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestral Reminiscence — Sorcery {3}{U} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Draw three cards, then discard a card."
//
// Rapturous Moment's loot without the mana: the three draws finish
// before the discard prompt opens, so a drawn card is a legal discard,
// and the caster picks which card to discard (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbfa2743-34ff-40e3-81f7-740fc08cd18a",
		Name:         "Ancestral Reminiscence",
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
