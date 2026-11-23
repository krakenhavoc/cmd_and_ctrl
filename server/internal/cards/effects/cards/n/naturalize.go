package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naturalize — Instant {1}{G}:
//
//	"Destroy target artifact or enchantment."
//
// The baseline the green disenchants are measured against — Krosan
// Grip costs one more for split second, Nature's Claim one less for
// giving the controller 4 life.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728",
		Name:         "Naturalize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
