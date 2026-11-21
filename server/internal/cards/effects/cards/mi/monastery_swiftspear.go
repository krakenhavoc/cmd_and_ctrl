package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monastery Swiftspear — Creature — Human Monk {R}, 1/2:
//
//	"Haste
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Both are canonical keywords the engine honours (game/prowess.go turns
// prowess into its trigger), so the card is its two printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dafd2713-d1bc-474b-b390-d2ff20b5375e",
		Name:            "Monastery Swiftspear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordProwess},
	})
}
