package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudfin Raptor — Creature — Bird Mutant {U}, 0/1 (Gatecrash, #5565):
//
//	"Flying
//	 Evolve (Whenever a creature you control enters, if that creature has
//	 greater power or toughness than this creature, put a +1/+1 counter on
//	 this creature.)"
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805);
// the other keyword is the engine's too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb0e1afa-fd9e-4bf9-bcb1-6e8110d6e2ff",
		Name:            "Cloudfin Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordEvolve},
	})
}
