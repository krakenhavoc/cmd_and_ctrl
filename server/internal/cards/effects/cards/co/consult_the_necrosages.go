package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consult the Necrosages — Sorcery {1}{U}{B} (Ravnica: City of Guilds,
// #5600):
//
//	"Choose one —
//	 • Target player draws two cards.
//	 • Target player discards two cards."
//
// Shredder's Revenge's two bullets without its life loss: each mode names
// its own target player, who may be the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9edddb4-0d60-4d21-8887-51d943c6a31f",
		Name:         "Consult the Necrosages",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeSteps("Target player draws two cards.",
				effects.TargetPlayer("target player"),
				effects.TmtB6ModeTargetDraws(2)),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2})),
			effects.ModeDoing("Target player discards two cards.",
				effects.TargetPlayer("target player"),
				effects.TheModesTargetDiscards(2)),
		),
	})
}
