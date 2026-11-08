package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Joraga Invocation — Sorcery {4}{G}{G} (Foundations, #4465):
//
//	"Each creature you control gets +3/+3 until end of turn and must be
//	 blocked this turn if able."
//
// Enlarge's pump and must-be-blocked requirement (BlockRequirementUntilEOT,
// CR 509.1c) on each creature you control. The set is fixed as the spell
// resolves (CR 611.2c, b23CreaturesYouControl), so a creature that enters
// later this turn gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "216d3eed-cb07-40c5-869a-962e074e757d",
		Name:         "Joraga Invocation",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, id := range b23CreaturesYouControl(ctx.Game, ctx.Controller()) {
				if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Joraga Invocation — +3/+3"}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
				if err := (BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementMustBeBlocked,
					Label:  "Joraga Invocation — must be blocked this turn",
				}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
