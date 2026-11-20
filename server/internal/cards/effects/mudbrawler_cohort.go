package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mudbrawler Cohort — Creature — Goblin Warrior {1}{R}, 1/1 (Shadowmoor,
// #5591):
//
//	"Haste
//	 This creature gets +1/+1 as long as you control another red
//	 creature."
//
// Haste is a printed keyword. shmB1Cohort: a layer 7c +1/+1 (CR 613.4c) on
// the Cohort alone, read continuously, while its controller controls
// ANOTHER creature of that colour (the Cohort itself never counts). The
// colour is a word of its text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "788bd9ae-b78e-4c86-bf12-d4e19780491a",
		Name:            "Mudbrawler Cohort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{
			shmB1Cohort("R", "This creature gets +1/+1 as long as you control another red creature."),
		},
	})
}
