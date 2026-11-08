package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider Food — Sorcery {2}{G} (Wilds of Eldraine, #4471):
//
//	"Destroy up to one target artifact, enchantment, or creature with
//	 flying. Create a Food token."
//
// The target is Broken Wings' (an artifact, an enchantment, or a creature
// with flying), with zero or one picked. With no target the spell still
// makes its Food; with one that became illegal it does nothing (CR
// 608.2b). The Food is the token table's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38eabfea-5ea5-4baa-8ad6-7a7cb00e1fed",
		Name:         "Spider Food",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("up to one target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return CreateToken{Controller: ctx.Controller(), Template: FoodToken(), N: 1}.Apply(ctx)
		},
	})
}
