package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Mechanic — Creature — Human Artificer {1}{U}, 2/1 (The Brothers'
// War, #5482):
//
//	"Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature.
//	 When this creature dies, create a 1/1 colorless Thopter artifact
//	 creature token with flying."
//
// Two printed abilities, two triggers. The token is made by the controller
// of the Mechanic as it died, whether or not it is still around to ask
// (CreateToken reads the trigger's controller).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7e7be75-646d-4c99-8388-15bb422c8e25",
		Name:         "Thopter Mechanic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82BroB07SecondDrawCounter("Thopter Mechanic"),
			effects.WhenThisDies("Thopter Mechanic — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
