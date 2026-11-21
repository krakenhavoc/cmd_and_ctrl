package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampaging Rendhorn — Creature — Beast {4}{G}, 4/4 (Ravnica Allegiance,
// #5521):
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)"
//
// Riot is the engine's keyword (ADR 0109 §10): the entry replacement asks
// for the counter or haste as it enters (CR 702.136a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b867deb4-a088-492e-95fe-487071ebfe8d",
		Name:            "Rampaging Rendhorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
	})
}
