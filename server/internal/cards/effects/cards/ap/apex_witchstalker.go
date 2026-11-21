package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apex Witchstalker — Creature — Wolf {4}{B}{B}, 6/4:
//
//	"Menace
//	 When this creature enters or dies, you gain 2 life.
//	 Basic landcycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "253d0cf7-65d3-4f13-aaf2-aae5bb60ac46",
		Name:            "Apex Witchstalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Apex Witchstalker — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
