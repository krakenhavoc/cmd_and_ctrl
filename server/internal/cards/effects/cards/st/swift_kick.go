package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Swift Kick — Instant {3}{R} (Jumpstart 2022, #5948):
//
//	"Target creature you control gets +1/+0 until end of turn. It fights
//	 target creature you don't control."
//
// Ambuscade's two clauses and resolution steps with a two-way fight: the
// boost goes on the creature you control if it is still legal, and the fight
// happens only while both targets are still legal (CR 701.14b, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c54fd75-e89f-490e-90dd-788f22b6e316",
		Name:         "Swift Kick",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Swift Kick — +1/+0 until end of turn"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SlotZeroFightsSlotOne(ctx.Game, ctx.Item)
			}),
		),
	})
}
