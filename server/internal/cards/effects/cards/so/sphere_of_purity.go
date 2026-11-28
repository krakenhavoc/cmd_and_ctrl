package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphere of Purity — Enchantment {3}{W} (Mirrodin, #5606):
//
//	"If an artifact would deal damage to you, prevent 1 of that damage."
//
// Protection of the Hekma's prevention (CR 615.1a) on damage to the controller
// alone, whose source is read as the damage snapshotted it (CR 608.2h) and must
// be an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "790d3ee2-8b93-464c-9b16-afe62db0048c",
		Name:         "Sphere of Purity",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB606PreventOneFromArtifactsToYou()},
	})
}
