package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adaptive Snapjaw — Creature — Lizard Beast {4}{G}, 6/2 (Gatecrash, #5565):
//
//	"Evolve (Whenever a creature you control enters, if that creature has
//	 greater power or toughness than this creature, put a +1/+1 counter on
//	 this creature.)"
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a56e2ad-684d-45a7-9145-479d1505bba2",
		Name:            "Adaptive Snapjaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
	})
}
