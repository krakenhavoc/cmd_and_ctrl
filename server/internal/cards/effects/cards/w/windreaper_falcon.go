package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Windreaper Falcon — Creature — Bird {1}{R}{G}, 1/1 (Mirage, S82 OLD b807):
//
//	"Flying, protection from blue"
//
// Flying and one protection quality (CR 702.16), carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa17782b-d8e7-423b-9f03-01ccb600f69a",
		Name:            "Windreaper Falcon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("U")},
	})
}
