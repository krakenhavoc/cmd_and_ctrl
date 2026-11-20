package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crabapple Cohort — Creature — Treefolk Warrior {4}{G}, 4/4 (Shadowmoor,
// #5591):
//
//	"This creature gets +1/+1 as long as you control another green
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
		OracleID:     "470f0043-3ffe-470b-981f-425f0583a87e",
		Name:         "Crabapple Cohort",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			shmB1Cohort("G", "This creature gets +1/+1 as long as you control another green creature."),
		},
	})
}
