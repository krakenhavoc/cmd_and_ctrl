package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purify — Sorcery {3}{W}{W} (Urza's Legacy, #5623):
//
//	"Destroy all artifacts and enchantments."
//
// Shatterstorm and Tranquility together: every artifact and every
// enchantment is destroyed at once (one event, so a death watcher sees them
// all), and an artifact creature goes as an artifact. Unlike Shatterstorm
// it says nothing about regeneration, so a regeneration shield saves one.
// Declared as a sweep of artifacts and enchantments.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfd1b276-ef5d-4f60-9f63-a548ad315e07",
		Name:         "Purify",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Or(effects.Artifact(), effects.Enchantment())}.Apply(ctx)
		},
	})
}
