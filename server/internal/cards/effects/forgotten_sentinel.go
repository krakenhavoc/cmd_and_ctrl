package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forgotten Sentinel — Artifact Creature — Golem {4}, 4/3 (Core Set 2021, #5511):
//
//	"This creature enters tapped."
//
// Diregraf Ghoul's enters-tapped replacement, and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9bd4a0bd-d10c-434a-8254-89f38740aafd",
		Name:         "Forgotten Sentinel",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}
