package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tome Scour — Sorcery {U} (Magic 2010, #5585):
//
//	"Target player mills five cards."
//
// Glimpse the Unthinkable's mill with five cards: the targeted player, if
// still legal as it resolves, mills five (CR 701.13a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "446a9d7e-9c1a-4e83-8342-2ed5acae2eed",
		Name:         "Tome Scour",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetPlayerMills(5)(ctx.Game, item)
		},
	})
}
