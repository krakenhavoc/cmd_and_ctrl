package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apex Witchstalker — Creature — Wolf {4}{B}{B}, 6/4:
//
//	"Menace
//	 When this creature enters or dies, you gain 2 life.
//	 Basic landcycling {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "253d0cf7-65d3-4f13-aaf2-aae5bb60ac46",
		Name:            "Apex Witchstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Apex Witchstalker — you gain 2 life", Do(GainLife{Amount: 2})),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
