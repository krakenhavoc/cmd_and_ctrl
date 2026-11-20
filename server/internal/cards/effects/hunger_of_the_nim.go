package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1af3c6ff-2884-4d8c-a01f-6f74d8ea10cc",
		Name:         "Hunger of the Nim",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := len(permanentsControlledByMatching(ctx.Game, ctx.Controller(), Artifact()))
			if n <= 0 {
				return nil
			}
			return BoostUntilEOT{Target: id, Power: n, Label: "Hunger of the Nim — +1/+0 for each artifact"}.Apply(ctx)
		},
	})
}
