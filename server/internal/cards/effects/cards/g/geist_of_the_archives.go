package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geist of the Archives — Creature — Spirit {2}{U}, 0/4 (Eldritch Moon,
// #5541):
//
//	"Defender
//	 At the beginning of your upkeep, scry 1."
//
// Defender is an engine keyword carried as printed data; the upkeep scry
// is Lifecrafter's Bestiary's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "345964ea-12b6-44be-b1a6-50178768b8f4",
		Name:            "Geist of the Archives",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Geist of the Archives — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
