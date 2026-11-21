package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hallowed Burial — Sorcery {3}{W}{W} (Eventide, #5590):
//
//	"Put all creatures on the bottom of their owners' libraries."
//
// Terminus without the miracle: every creature goes at once, and each
// owner with more than one orders their own (CR 401.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcaaddee-00b1-495c-9161-40c91c900aac",
		Name:         "Hallowed Burial",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TuckAllCreaturesToBottomByOwner(ctx, item, "Hallowed Burial — put these on the bottom of your library in any order")
		},
	})
}
