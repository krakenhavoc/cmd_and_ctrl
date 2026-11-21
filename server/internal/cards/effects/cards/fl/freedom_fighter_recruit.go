package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freedom Fighter Recruit — Creature — Human Rebel Ally {1}{R}, */2
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Freedom Fighter Recruit's power is equal to the number of creatures
//	 you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a), Adeline's and
// Esgaroth Garrison's shape: it counts the Recruit itself, and it functions
// in all zones (the ruling), which the engine's anywhere reading of a `*`
// gives it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33e66662-aea6-480c-b69b-1907d987b69e",
		Name:         "Freedom Fighter Recruit",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AvB30aPowerEqualsCreaturesYouControl("Power is equal to the number of creatures you control."),
		},
	})
}
