package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frost Bite — Snow Instant {R} (Kaldheim):
//
//	"Frost Bite deals 2 damage to target creature or planeswalker. If you
//	 control three or more snow permanents, it deals 3 damage instead."
//
// The snow permanents are counted as the spell resolves, so the Frost Bite
// itself (not a permanent) never counts and one that arrives or leaves in
// response changes the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "128170ed-c86a-4f02-9244-28197be90c10",
		Name:         "Frost Bite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			amount := 2
			if effects.S82OldB064SnowPermanentsYouControl(ctx.Game, item.Controller) >= 3 {
				amount = 3
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: amount}.Apply(ctx)
		},
	})
}
