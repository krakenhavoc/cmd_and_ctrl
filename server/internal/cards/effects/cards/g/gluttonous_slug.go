package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gluttonous Slug — Creature — Slug Horror {1}{B}, 0/3 (Modern Horizons, #5519):
//
//	"Menace
//	 Evolve (Whenever a creature you control enters, if that creature has
//	 greater power or toughness than this creature, put a +1/+1 counter on
//	 this creature.)"
//
// Menace and the engine's evolve keyword (CR 702.100), Elusive Krasis's
// shape: the comparison is made on entry and again as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68fbc13c-00b3-49bb-b343-4f2f1e8f5a7c",
		Name:            "Gluttonous Slug",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", game.KeywordEvolve},
	})
}
