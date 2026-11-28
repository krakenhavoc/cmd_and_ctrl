package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rock Slide — Instant {X}{R} (Visions, S82 OLD batch 786, #5631):
//
//	"Rock Slide deals X damage divided as you choose among any number of
//	 target attacking or blocking creatures without flying."
//
// Hail of Arrows' shape (any number of targets, X divided, CR 601.2d) over
// the attacking or blocking creatures that do not have flying. Every pick
// must be one as the spell is cast and again as it resolves; a target that
// has stopped being legal takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b2e7c1a-dc07-4138-a560-dc80fa723860",
		Name:         "Rock Slide",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets: effects.TargetCreature("any number of target attacking or blocking creatures without flying",
			effects.AttackingOrBlocking(), effects.WithoutKeyword("flying")).WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
