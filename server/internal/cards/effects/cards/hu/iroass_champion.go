package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iroas's Champion — Creature — Human Soldier {1}{R}{W}, 2/2 (Magic Origins, #5549):
//
//	"Double strike (This creature deals both first-strike and regular combat damage.)"
//
// Boros Swiftblade's shape: double strike is the engine's two damage
// steps (CR 702.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f2252f6-69b9-4c56-b078-803db301a874",
		Name:            "Iroas's Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
