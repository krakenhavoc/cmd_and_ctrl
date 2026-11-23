package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "077ae62d-ed94-4365-a8b7-97fd99cff028",
		Name:         "Rabid Gnaw",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
		},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if biter, ok := ctx.ClauseTarget(0); ok && biter.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: biter.ID, Power: 1, Label: "Rabid Gnaw — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				// The bite reads the biter's power as it now stands.
				ctx.Game.RecomputeLayersIfStaleLocked()
			}
			return effects.OneSidedBite(item, ctx)
		},
	})
}
