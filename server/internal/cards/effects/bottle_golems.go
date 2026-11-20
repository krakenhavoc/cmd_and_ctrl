package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bottle Golems — Artifact Creature — Golem {4}, 3/3 (Modern Horizons 2,
// #5501):
//
//	"Trample
//	 When this creature dies, you gain life equal to its power."
//
// Willow Geist's dies trigger: "its power" is the Golems' power as they
// last existed on the battlefield (CR 603.10a, 608.2h), counters and
// continuous effects included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "46334ae3-e496-4f15-8033-0f42ee0ae375",
		Name:            "Bottle Golems",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Bottle Golems — gain life equal to its power",
				Do(GainLifeEqualToSubject{Stat: SubjectPower})),
		},
	})
}
