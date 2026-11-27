package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Last Caress — Sorcery {2}{B} (Apocalypse, S82 OLD, #5614):
//
//	"Target player loses 1 life and you gain 1 life.
//	 Draw a card."
//
// No simplification.
func init() {
	drain := effects.TargetPlayerLosesLifeYouGain(1)
	effects.Register(effects.Spec{
		OracleID:     "620b13cf-b8ea-48b7-8300-3c48390e32e3",
		Name:         "Last Caress",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := drain(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
