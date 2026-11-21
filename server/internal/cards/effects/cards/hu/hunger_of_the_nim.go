package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunger of the Nim — Sorcery {1}{B} (Darksteel, #5605):
//
//	"Target creature gets +1/+0 until end of turn for each artifact you
//	 control."
//
// Giant Growth's boost with the amount counted as the spell resolves (CR
// 608.2h): the artifacts its caster controls then. The bonus is locked in
// (CR 611.2c), so losing an artifact later doesn't shrink it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1af3c6ff-2884-4d8c-a01f-6f74d8ea10cc",
		Name:         "Hunger of the Nim",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Artifact()))
			if n <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: n, Label: "Hunger of the Nim — +1/+0 for each artifact"}.Apply(ctx)
		},
	})
}
