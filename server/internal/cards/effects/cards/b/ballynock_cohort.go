package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ballynock Cohort — Creature — Kithkin Soldier {2}{W}, 2/2 (Shadowmoor,
// #5591):
//
//	"First strike
//	 This creature gets +1/+1 as long as you control another white
//	 creature."
//
// First strike is a printed keyword. shmB1Cohort: a layer 7c +1/+1 (CR
// 613.4c) on the Cohort alone, read continuously, while its controller
// controls ANOTHER creature of that colour (the Cohort itself never
// counts). The colour is a word of its text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7c415180-c179-474b-8c58-e7096bd2c763",
		Name:            "Ballynock Cohort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			effects.ShmB1Cohort("W", "This creature gets +1/+1 as long as you control another white creature."),
		},
	})
}
