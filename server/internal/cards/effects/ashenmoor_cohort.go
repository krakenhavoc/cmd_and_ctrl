package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashenmoor Cohort — Creature — Elemental Warrior {5}{B}, 4/3 (Shadowmoor,
// #5591):
//
//	"This creature gets +1/+1 as long as you control another black
//	 creature."
//
// shmB1Cohort: a layer 7c +1/+1 (CR 613.4c) on the Cohort alone, read
// continuously, while its controller controls ANOTHER creature of that
// colour (the Cohort itself never counts). The colour is a word of its
// text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c627f1a8-9a57-4eb9-a8a5-2c285446fe97",
		Name:         "Ashenmoor Cohort",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			shmB1Cohort("B", "This creature gets +1/+1 as long as you control another black creature."),
		},
	})
}
