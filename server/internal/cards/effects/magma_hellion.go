package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magma Hellion — Creature — Hellion {6}{R}:
//
//	"Assist (Another player can pay up to {6} of this spell's cost.)
//	 Trample, haste"
//
// A vanilla pair of keywords with assist. Assist is the printed keyword
// the engine reads off the card (CR 702.132a, ADR 0150 §4); trample and
// haste are plain printed keywords.
func init() {
	Register(Spec{
		OracleID:        "88f6575d-a6cb-4b2c-90dd-c6239f3c6e4d",
		Name:            "Magma Hellion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "haste", "trample"},
	})
}
