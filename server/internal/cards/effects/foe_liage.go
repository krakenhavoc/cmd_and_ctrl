package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foe-liage — {3}{G} Creature — Plant Mutant 3/3 (Teenage Mutant Ninja
// Turtles Commander, #5011):
//
//	"Whenever a land enters during your turn, put a +1/+1 counter on this
//	 creature."
//
// Any land, under any player's control, so long as it enters during
// Foe-liage's controller's turn: one trigger per land. A Foe-liage that has
// left by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "017cadd8-55ee-497a-84d4-6535173bc1eb",
		Name:         "Foe-liage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, tmcB8ALandEnteredDuringYourTurn,
				"Foe-liage — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
