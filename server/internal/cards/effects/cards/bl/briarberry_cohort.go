package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Briarberry Cohort — Creature — Faerie Soldier {1}{U}, 1/1 (Shadowmoor,
// #5591):
//
//	"Flying
//	 This creature gets +1/+1 as long as you control another blue
//	 creature."
//
// Flying is a printed keyword. shmB1Cohort: a layer 7c +1/+1 (CR 613.4c)
// on the Cohort alone, read continuously, while its controller controls
// ANOTHER creature of that colour (the Cohort itself never counts). The
// colour is a word of its text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "598283f3-ae29-43b9-8743-6c065136dcad",
		Name:            "Briarberry Cohort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.ShmB1Cohort("U", "This creature gets +1/+1 as long as you control another blue creature."),
		},
	})
}
