package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diregraf Ghoul — Creature — Zombie {B}, 2/2 (Foundations, #4465):
//
//	"This creature enters tapped."
//
// Alley Assailant's enters-tapped replacement, and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6048fc70-0dcc-4b54-977d-16e240225f82",
		Name:         "Diregraf Ghoul",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}
