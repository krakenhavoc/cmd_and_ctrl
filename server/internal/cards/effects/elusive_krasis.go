package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elusive Krasis — Creature — Fish Mutant {1}{G}{U}, 0/4:
//
//	"This creature can't be blocked.
//	 Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)"
//
// "Can't be blocked" is the S24 restriction bit on the creature itself,
// read at the block declaration (CR 509.1b) — Blighted Agent's shape.
// Evolve is the engine's keyword trigger (game/evolve.go, #1805).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b876101b-7b18-48d7-9c8c-aaa6b1b5c199",
		Name:            "Elusive Krasis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
