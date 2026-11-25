package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tide Drifter — Creature — Eldrazi Drone {1}{U}, 0/5 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Other colorless creatures you control get +0/+1."
//
// Kozilek, the Broken Reality's anthem at +0/+1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "428c18aa-b195-4cf2-a1a9-09723b26eb0a",
		Name:            "Tide Drifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Static:          []game.StaticAbility{effects.OldB305OtherColorlessAnthem(0, 1, "Other colorless creatures you control get +0/+1.")},
	})
}
