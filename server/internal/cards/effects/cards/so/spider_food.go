package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "38eabfea-5ea5-4baa-8ad6-7a7cb00e1fed",
		Name:         "Spider Food",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("up to one target artifact, enchantment, or creature with flying",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.And(effects.Creature(), effects.HasKeyword("flying")))).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.FoodToken(), N: 1}.Apply(ctx)
		},
	})
}
