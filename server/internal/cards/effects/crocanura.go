package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crocanura — Creature — Crocodile Frog {2}{G}, 1/3 (Gatecrash, #5565):
//
//	"Reach (This creature can block creatures with flying.)
//	 Evolve (Whenever a creature you control enters, if that creature has
//	 greater power or toughness than this creature, put a +1/+1 counter on
//	 this creature.)"
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805);
// the other keyword is the engine's too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2071f278-77ba-46f8-85a0-ccd761199f29",
		Name:            "Crocanura",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", game.KeywordEvolve},
	})
}
