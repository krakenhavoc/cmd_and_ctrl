package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tear Asunder — {1}{G} Instant:
//
//	"Kicker {1}{B} (You may pay an additional {1}{B} as you cast this
//	 spell.)
//	 Exile target artifact or enchantment. If this spell was kicked,
//	 exile target nonland permanent instead."
//
// The kicked clause replaces the printed one (#1716, WhenPaid): a
// creature is a legal target only for a kicked cast, at announce and
// again at resolution (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "610af0f7-b5e3-43fb-9d02-7c59bd99034c",
		Name:         "Tear Asunder",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{1}{B}"),
			effects.TargetPermanent("target nonland permanent", effects.Nonland()))},
		Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.ExileTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
