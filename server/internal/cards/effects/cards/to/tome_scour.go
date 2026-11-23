package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tome Scour — Sorcery {U} (Magic 2010, #5585):
//
//	"Target player mills five cards."
//
// Glimpse the Unthinkable's mill with five cards: the targeted player, if
// still legal as it resolves, mills five (CR 701.13a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "446a9d7e-9c1a-4e83-8342-2ed5acae2eed",
		Name:         "Tome Scour",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerMillsEffect(5)(ctx.Game, item)
		},
	})
}
