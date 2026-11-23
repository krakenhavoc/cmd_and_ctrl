package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli's Final Strike — Instant {2}{G} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Target creature you control gets +1/+0 until end of turn. It deals
//	 damage equal to its power to target creature an opponent controls."
//
// Knockout Maneuver's one-sided bite with a +1/+0 in place of the
// counter: the pump lands first, and the bite (oneSidedBite) reads the
// creature's power afterwards, so it hits for the pumped amount. If the
// creature you control is no longer a legal target, there is no pump and
// no damage; if only the opponent's creature is gone, the pump still
// happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20acb8df-2ea0-48b0-8dcf-54dd0e1f8db8",
		Name:         "Huatli's Final Strike",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Huatli's Final Strike — +1/+0 until end of turn"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.OneSidedBite(ctx.Item, ctx) }),
		),
	})
}
