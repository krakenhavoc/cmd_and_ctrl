package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deputized Protester — Creature — Human Warrior {2}{R}, 2/1 (Ravnica:
// Clue Edition, bonus reprint of Murders at Karlov Manor):
//
//	"Menace
//	 Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)"
//
// Both are engine keywords (CR 702.111, 702.121).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09947592-1730-4c1c-bfa5-be53a68d6919",
		Name:            "Deputized Protester",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", game.KeywordMelee},
	})
}
