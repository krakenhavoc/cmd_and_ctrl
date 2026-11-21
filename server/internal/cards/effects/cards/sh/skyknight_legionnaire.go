package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyknight Legionnaire — Creature — Human Knight {1}{R}{W}, 2/2 (Ravnica:
// City of Guilds, #5600):
//
//	"Flying, haste"
//
// Flying and haste are engine keywords carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61178a6c-70b7-447d-87ee-a8d9369c3d15",
		Name:            "Skyknight Legionnaire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
