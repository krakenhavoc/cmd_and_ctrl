package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Supply-Line Cranes — Creature — Bird {3}{W}{W}, 2/4 (Journey into Nyx, #5558):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on target creature."
//
// Ironpaw Aspirant's enters counter on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bf3811b3-f3d2-4ffb-8019-e403a2cc1d81",
		Name:            "Supply-Line Cranes",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Supply-Line Cranes — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
	})
}
