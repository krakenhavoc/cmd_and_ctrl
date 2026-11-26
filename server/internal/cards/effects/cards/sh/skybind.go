package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skybind — Enchantment {3}{W}{W} (Journey into Nyx, S82 OLD, #5558):
//
//	"Constellation — Whenever this enchantment or another enchantment you
//	 control enters, exile target nonenchantment permanent. Return that card to
//	 the battlefield under its owner's control at the beginning of the next end
//	 step."
//
// Vizier of Deferment's exile and shared next-end-step return, on the
// constellation condition (Agent of Erebos). A token exiled this way ceases to
// exist and is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13d34f21-f0ce-4e59-ac3e-6635bf7292d3",
		Name:         "Skybind",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.ConstellationTriggers, "Skybind — exile target nonenchantment permanent; return it at the beginning of the next end step (constellation)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Skybind — return the exiled card")
				}),
				effects.TargetPermanent("target nonenchantment permanent", effects.Not(effects.Enchantment()))),
		},
	})
}
