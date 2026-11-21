package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mugato — {3}{G}{G} Creature — Ape Beast, 6/6 (Star Trek, #4973):
// "Reach, vigilance, deathtouch".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb9a3b52-62ad-467f-b62a-0273ceee48e5",
		Name:            "Mugato",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "deathtouch"},
	})
}
