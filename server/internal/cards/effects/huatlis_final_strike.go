package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "20acb8df-2ea0-48b0-8dcf-54dd0e1f8db8",
		Name:         "Huatli's Final Strike",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerRemove, // a one-sided fight, as Bite Down (#4734)
			Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0))},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Huatli's Final Strike — +1/+0 until end of turn"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
		),
	})
}
