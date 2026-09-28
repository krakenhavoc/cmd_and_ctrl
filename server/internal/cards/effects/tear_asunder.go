package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "610af0f7-b5e3-43fb-9d02-7c59bd99034c",
		Name:         "Tear Asunder",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{1}{B}"),
			TargetPermanent("target nonland permanent", Nonland()))},
		Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return ExileTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
