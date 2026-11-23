package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zhur-Taa Goblin — Creature — Goblin Berserker {R}{G}, 2/2 (Ravnica
// Allegiance, #5521):
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
		OracleID:        "1003304f-af01-4235-af82-6ef901bd00bf",
		Name:            "Zhur-Taa Goblin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
	})
}
