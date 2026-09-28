package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talruum Piper — Creature — Minotaur, {4}{R}, 3/3:
//
//	"All creatures with flying able to block this creature do so."
//
// #1684: the filtered Lure — FilteredLure(game.BlockerFilterFlying).
// The filter reads the would-be blocker's flying at the declaration, so
// a creature that lost flying is free and one that gained it is bound.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d18e60ed-ab6e-4b73-88f6-caa3f8e9c78e",
		Name:         "Talruum Piper",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{FilteredLure(game.BlockerFilterFlying)},
	})
}
