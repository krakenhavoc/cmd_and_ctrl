package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c9edddb4-0d60-4d21-8887-51d943c6a31f",
		Name:         "Consult the Necrosages",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeSteps("Target player draws two cards.",
				TargetPlayer("target player"),
				tmtB6ModeTargetDraws(2)),
				ForTargets(game.TargetPurpose{Slot: 0, Draws: 2})),
			ModeDoing("Target player discards two cards.",
				TargetPlayer("target player"),
				TheModesTargetDiscards(2)),
		),
	})
}
