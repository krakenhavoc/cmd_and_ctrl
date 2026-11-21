package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "017cadd8-55ee-497a-84d4-6535173bc1eb",
		Name:         "Foe-liage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.TmcB8ALandEnteredDuringYourTurn,
				"Foe-liage — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
