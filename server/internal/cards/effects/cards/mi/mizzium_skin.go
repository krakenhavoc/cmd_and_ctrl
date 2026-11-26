package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzium Skin — Instant {U} (Return to Ravnica, #5566):
//
//	"Target creature you control gets +0/+1 and gains hexproof until end of turn.
//	 Overload {1}{U} (You may cast this spell for its overload cost. If you do,
//	 change "target" in its text to "each.")"
//
// Overloaded it names no target (CR 702.96b) and gives each creature you control
// at resolution +0/+1 and hexproof until end of turn (CR 611.2c). Cast normally
// the one target must still be legal on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02d998a6-0afd-4be1-a209-05618fcc43ea",
		Name:         "Mizzium Skin",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 1, "hexproof")),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{1}{U}"), game.Purpose{Answers: game.AnswerPump | game.AnswerProtect}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, match, ok := effects.OldB424Subject(ctx, effects.YouControl())
			if !ok {
				return nil
			}
			if err := effects.UntilEndOfTurn(ctx, id, match, "Mizzium Skin — +0/+1", game.ModifyPTMod(0, 1)); err != nil {
				return err
			}
			return effects.UntilEndOfTurn(ctx, id, match, "Mizzium Skin — hexproof", game.AddKeywordsMod("hexproof"))
		},
	})
}
