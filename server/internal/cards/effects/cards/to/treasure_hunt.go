package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treasure Hunt — Sorcery {1}{U} (Commander 2018, S82 OLD batch 193, #5526):
//
//	"Reveal cards from the top of your library until you reveal a nonland
//	 card, then put all cards revealed this way into your hand."
//
// The whole run is revealed at once (CR 701.20a) and every revealed card,
// the nonland hit and the lands before it, goes to your hand. A library with
// no nonland card reveals itself whole and all of it comes to your hand;
// revealing is not drawing, so an empty library is not a loss. A token in the
// library never counts as the nonland card (CR 108.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05079479-86a6-4041-a395-83d325b6ddb7",
		Name:         "Treasure Hunt",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			run, _ := effects.RevealUntil(ctx, item.Controller, func(c game.Card) bool { return !c.IsLand() },
				"Treasure Hunt — revealed until a nonland card")
			ctx.Game.TakeFromLibraryToHandForEffect(item.Controller, run)
			return nil
		},
	})
}
