package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vinelasher Kudzu — Creature — Plant {1}{G}, 1/1 (Ravnica: City of Guilds,
// #5600):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 this creature."
//
// Territorial Scythecat's landfall counter, without the trample.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62204fac-a367-475c-bfb9-80a87af2d2f1",
		Name:         "Vinelasher Kudzu",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Vinelasher Kudzu — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
