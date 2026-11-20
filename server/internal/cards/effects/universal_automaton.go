package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Universal Automaton — Artifact Creature — Shapeshifter {1}, 1/1 (Modern Horizons, #5519):
//
//	"Changeling (This card is every creature type.)"
//
// A colourless one-drop changeling: every creature type in every zone
// (CR 702.73a), and an artifact as well.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d50b1d2e-1825-432a-ab73-068fdc356f50",
		Name:            "Universal Automaton",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
	})
}
