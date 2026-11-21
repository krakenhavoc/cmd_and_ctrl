package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox Charger — Creature — Rhino Soldier {3}{G}, 3/3 (Shards of Alara,
// #5588):
//
//	"Trample
//	 Exalted"
//
// Both are printed keywords; exalted's trigger is derived by the engine
// (CR 702.83a, game/exalted.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1b7db10-58a1-46fc-ac6c-287393ed96c2",
		Name:            "Rhox Charger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordExalted},
	})
}
