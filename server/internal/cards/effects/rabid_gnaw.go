package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabid Gnaw — Instant {1}{R} (Bloomburrow, #4467):
//
//	"Target creature you control gets +1/+0 until end of turn. Then it
//	 deals damage equal to its power to target creature you don't
//	 control."
//
// Wolf Strike without the night condition: the boost first, then Bite
// Down's one-sided fight (oneSidedBite) reading the biter's power after
// the boost, which is the printed order. If your creature has left, it
// gets nothing and deals nothing; if only the other creature has left,
// yours still gets +1/+0 (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "077ae62d-ed94-4365-a8b7-97fd99cff028",
		Name:         "Rabid Gnaw",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0)),
		},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if biter, ok := ctx.ClauseTarget(0); ok && biter.Kind == game.TargetCard {
				if err := (BoostUntilEOT{Target: biter.ID, Power: 1, Label: "Rabid Gnaw — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				// The bite reads the biter's power as it now stands.
				ctx.Game.RecomputeLayersIfStaleLocked()
			}
			return oneSidedBite(item, ctx)
		},
	})
}
