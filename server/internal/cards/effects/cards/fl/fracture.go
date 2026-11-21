package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fracture — Instant {W}{B} (EDHREC rank 1256):
//
//	"Destroy target artifact, enchantment, or planeswalker."
//
// Mortify's shape with a three-way target clause; Despark's
// colours. The target predicate is the whole card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f21d0319-0509-4ac1-b6e3-10955a26fd7a",
		Name:         "Fracture",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, enchantment, or planeswalker",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
