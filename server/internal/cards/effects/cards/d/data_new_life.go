package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Data, New Life — Legendary Artifact Creature — Android Scientist {1}{U},
// 0/1 (Star Trek, #4973):
//
//	"Whenever you draw a card, put a +1/+1 counter on Data."
//
// Chasm Skulker's draw trigger: one counter per card its controller draws,
// on any player's turn; an opponent's draw does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "010630c6-5d7b-43e0-90a9-3e3cef9a4e6b",
		Name:         "Data, New Life",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Data, New Life — put a +1/+1 counter on Data",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
