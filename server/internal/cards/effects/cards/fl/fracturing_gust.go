package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fracturing Gust — Instant {2}{G/W}{G/W}{G/W} (Secret Lair Drop):
//
//	"Destroy all artifacts and enchantments. You gain 2 life for each
//	 permanent destroyed this way."
//
// DestroyAllMatching over artifacts and enchantments with the life as its
// tail, told only the permanents that were really destroyed: an indestructible
// one, or one a regeneration shield saved, gains nothing (Culling Ritual's
// shape). Tokens count; so do the caster's own permanents.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "011d4f47-c6fd-434b-98f2-2f558fbff40b",
		Name:         "Fracturing Gust",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}},
		OnResolve:    effects.S82OldB023DestroyArtifactsAndEnchantments,
	})
}
