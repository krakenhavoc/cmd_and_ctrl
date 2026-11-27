package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dynacharge — Instant {R} (Return to Ravnica, #5566):
//
//	"Target creature you control gets +2/+0 until end of turn.
//	 Overload {2}{R} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// Overloaded it names no target (CR 702.96b) and pumps each creature you control
// at resolution (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61869f52-51a8-4126-af06-5f5c3d80c28c",
		Name:         "Dynacharge",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{2}{R}"), game.Purpose{Answers: game.AnswerPump}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, match, ok := effects.OldB424Subject(ctx, effects.YouControl())
			if !ok {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, match, "Dynacharge — +2/+0", game.ModifyPTMod(2, 0))
		},
	})
}
