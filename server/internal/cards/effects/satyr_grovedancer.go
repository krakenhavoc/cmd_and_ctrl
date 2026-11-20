package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Satyr Grovedancer — Creature — Satyr Shaman {1}{G}, 1/1 (Journey into Nyx, #5558):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Ironpaw Aspirant's enters counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49f3bd5a-9c2a-45b8-aa10-9c6471028a0f",
		Name:         "Satyr Grovedancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Satyr Grovedancer — put a +1/+1 counter on target creature", PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
	})
}
