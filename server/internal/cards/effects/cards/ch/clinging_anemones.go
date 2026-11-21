package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clinging Anemones — Creature — Jellyfish {3}{U}, 1/4 (Gatecrash, #5565):
//
//	"Defender
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
		OracleID:        "7058e7c3-b704-47f6-85f8-ee584f76a1fe",
		Name:            "Clinging Anemones",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", game.KeywordEvolve},
	})
}
