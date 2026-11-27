package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eye Spy — {U} Sorcery (Starter 1999, S82 OLD batch 701, #5620):
//
//	"Look at the top card of target player's library. You may put that card
//	 into their graveyard."
//
// Only the controller looks (CR 701.20); the pick is optional
// (S82OldB701EyeSpy).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fef3f4f-18b6-452c-a634-c870709abcf5",
		Name:         "Eye Spy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB701EyeSpy(ctx.Game, item)
		},
	})
}
