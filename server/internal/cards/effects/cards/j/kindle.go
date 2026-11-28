package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kindle — Instant {1}{R} (Tempest, S82 OLD batch 757, #5628):
//
//	"Kindle deals X damage to any target, where X is 2 plus the number of
//	 cards named Kindle in all graveyards."
//
// X is counted as the spell resolves (CR 608.2h); Kindle itself is still on
// the stack then, so it is not counted. Names compare as every card name
// does (CR 201.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9397a23b-770a-43fc-8bc2-2c3abb0683c5",
		Name:         "Kindle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			amount := 2 + effects.S82OldB757CardsNamedInGraveyards(ctx.Game, "Kindle")
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: amount}.Apply(ctx)
		},
	})
}
