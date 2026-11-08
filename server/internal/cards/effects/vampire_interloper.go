package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Interloper — Creature — Vampire Scout {1}{B}, 2/1
// (Foundations, #4465):
//
//	"Flying
//	 This creature can't block."
//
// Flying from PrintedKeywords, and Carrion Feeder's "can't block" as a
// RestrictSelf static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b753d7dd-4c6b-4472-a009-671572f927eb",
		Name:            "Vampire Interloper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
