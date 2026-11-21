package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Makeshift Battalion — Creature — Human Soldier {2}{W}, 3/2 (Ravnica
// Remastered, #5948):
//
//	"Battalion — Whenever this creature and at least two other creatures
//	 attack, put a +1/+1 counter on this creature."
//
// Battalion is Boros Strike-Captain's group test (gtcBattalion); the counter
// goes on this creature only while it is still the permanent the ability came
// from (CR 400.7), through the counter window.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b6e2feb-9e8f-4912-ada0-1614502d9813",
		Name:         "Makeshift Battalion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gtcBattalion("Makeshift Battalion — put a +1/+1 counter on it (battalion)",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
