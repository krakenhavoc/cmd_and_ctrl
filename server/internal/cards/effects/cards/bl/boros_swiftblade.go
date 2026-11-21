package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boros Swiftblade — Creature — Human Soldier {R}{W}, 1/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Double strike"
//
// The printed keyword (CR 702.4): it deals both first-strike and regular
// combat damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "377e4dad-866b-4e02-8c10-072bf7f000be",
		Name:            "Boros Swiftblade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
