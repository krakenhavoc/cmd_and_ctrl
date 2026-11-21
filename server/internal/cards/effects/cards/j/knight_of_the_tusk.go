package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of the Tusk — Creature — Human Knight {4}{W}{W}, 3/7
// (Core Set 2019, #5527):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)"
//
// The whole text is printed keywords; the engine reads them off
// PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "524d81fb-33f5-4e77-ae41-477b91e67be4",
		Name:            "Knight of the Tusk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
