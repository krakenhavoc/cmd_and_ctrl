package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Your Temple Is Under Attack — Instant {2}{W} (Commander Legends: Battle
// for Baldur's Gate, #5490):
//
//	"Choose one —
//	 • Pray for Protection — Creatures you control gain indestructible
//	   until end of turn.
//	 • Strike a Deal — You and target opponent each draw two cards."
//
// Pray for Protection is Duty Beyond Death's grant: the creatures you
// control as the spell resolves gain indestructible, and one that arrives
// later this turn does not (CR 611.2c). Strike a Deal is Commander William
// Riker's draw on a spell's bullet: you draw two, then the opponent it
// targets draws two (Ashling's Command's TheModesTargetPlayerDrawsTwo),
// the second draw waiting for any prompt the first asks (ModeSteps, CR
// 608.2c). If that opponent is no longer a legal
// target the spell does nothing, so you draw nothing either (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f48f69c0-cca8-4f44-84cc-71e7cbbb9208",
		Name:         "Your Temple Is Under Attack",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeSteps("Pray for Protection — Creatures you control gain indestructible until end of turn.", nil,
				effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Pray for Protection — indestructible",
				}), game.Purpose{Answers: game.AnswerSurvive}),
			effects.ModeWithPurpose(effects.ModeSteps("Strike a Deal — You and target opponent each draw two cards.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.DrawCards{N: 2},
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.TheModesTargetPlayerDrawsTwo(ctx.Item, ctx, ctx.Occurrence())
				})),
				game.Purpose{Draws: 2, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2})}),
		),
	})
}
