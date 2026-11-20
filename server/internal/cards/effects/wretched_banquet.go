package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wretched Banquet — Sorcery {B} (Conflux, #5587):
//
//	"Destroy target creature if it has the least power or is tied for
//	 least power among creatures on the battlefield."
//
// Drop of Honey's least-power reading (least_power.go) as a check on the
// target: every creature on the battlefield as the spell resolves,
// anyone's, compared after the layers (CR 608.2h). A target with more
// power than the least is not destroyed, and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dafd4385-e64f-4a4a-93c4-b0ac6bd30d41",
		Name:         "Wretched Banquet",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			for _, least := range creaturesWithLeastPower(ctx.Game) {
				if least == id {
					return DestroyTarget{Target: id}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
