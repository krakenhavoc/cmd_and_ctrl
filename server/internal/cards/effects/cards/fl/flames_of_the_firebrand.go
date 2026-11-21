package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flames of the Firebrand — Sorcery {2}{R} (#1658, unblocked by
// #1656):
//
//	"Flames of the Firebrand deals 3 damage divided as you choose
//	 among one, two, or three targets."
//
// Arc Lightning under a different name — same fixed-3, one-to-three
// shape, same CR 601.2d division rule.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc32c24f-2f2a-4125-917c-60d425166640",
		Name:         "Flames of the Firebrand",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
