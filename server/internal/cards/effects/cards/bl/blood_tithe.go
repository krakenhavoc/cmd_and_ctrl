package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Tithe — Sorcery {3}{B} (Magic 2011, #5579):
//
//	"Each opponent loses 3 life. You gain life equal to the life lost
//	 this way."
//
// Kokusho's drain at a fixed 3 (b21DrainEachOpponentAndGainTheTotal):
// life loss, not damage, and the gain is what the opponents really lost,
// so three opponents at full life give 9.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c82665b7-d011-4bdc-b4a2-eca5861b16d2",
		Name:         "Blood Tithe",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B21DrainEachOpponentAndGainTheTotal(ctx.Game, item, 3)
		},
	})
}
