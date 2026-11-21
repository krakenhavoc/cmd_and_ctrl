package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosemane Centaur — Creature — Centaur Soldier {3}{G}{W}, 4/4 (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 Vigilance"
//
// Sun-Dappled Celebrant's shape: convoke is the shared tap cost (CR
// 702.51a) and vigilance is printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54b3ccdc-e5b0-44d9-bc81-a296fc5dc005",
		Name:            "Rosemane Centaur",
		Completeness:    effects.CompletenessFull,
		TapCost:         effects.Convoke(),
		PrintedKeywords: []string{"vigilance"},
	})
}
