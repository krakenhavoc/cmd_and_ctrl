package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disenchant — Instant {1}{W} (EDHREC rank 1103):
//
//	"Destroy target artifact or enchantment."
//
// The original. Naturalize in white: the same target clause and the
// same single primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7e97fa9-4b72-4548-b854-5be5f18a6f1a",
		Name:         "Disenchant",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
