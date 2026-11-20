package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shambling Ghoul — Creature — Zombie {1}{B}, 2/3 (Magic Origins, #5549):
//
//	"This creature enters tapped."
//
// Diregraf Ghoul's enters-tapped replacement (CR 614.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2acc1657-d716-41d7-aedf-9b080a58ec83",
		Name:         "Shambling Ghoul",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}
