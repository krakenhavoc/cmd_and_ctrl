package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charging Binox — Creature — Beast {7}{G}:
//
//	"Assist (Another player can pay up to {7} of this spell's cost.)
//	 Trample"
//
// A vanilla trampler with assist. Assist is the printed keyword
// (CR 702.132a, ADR 0150 §4): the engine reads it off the card
// (game.HasAssist), and the cast waits for the helper's share of the
// generic mana. Nothing else is declared.
func init() {
	Register(Spec{
		OracleID:        "0f10f9d1-f9f9-4356-9208-0f21f44bfb97",
		Name:            "Charging Binox",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "trample"},
	})
}
