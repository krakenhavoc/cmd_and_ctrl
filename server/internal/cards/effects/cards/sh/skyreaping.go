package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyreaping — Sorcery {1}{G} (Born of the Gods, S82 OLD b378, #5559):
//
//	"Skyreaping deals damage to each creature with flying equal to your
//	 devotion to green."
//
// Devotion is counted once as the spell resolves (CR 700.5); the damage goes
// to every creature with flying, yours too, through DamageEachMatching.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcdc60ae-1858-4b51-8b64-58a4942bc3d7",
		Name:         "Skyreaping",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.S82OldB378DevotionOf(ctx, "G")
			if n <= 0 {
				return nil
			}
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), n)
		},
	})
}
