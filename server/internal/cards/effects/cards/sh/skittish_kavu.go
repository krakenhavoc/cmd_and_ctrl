package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skittish Kavu — Creature — Kavu {1}{R}, 1/1 (Invasion, S82 OLD #5616):
//
//	"This creature gets +1/+1 as long as no opponent controls a white or blue
//	 creature."
//
// A layer 7c static read off the table as it stands (CR 604.2): an opponent's
// creature that is white or blue, in either colour or both, through the layers
// (CR 105.2) turns it off, and it comes back when the last one goes. A white or
// blue creature that you control does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55722e42-a729-4289-895f-be913398c434",
		Name:         "Skittish Kavu",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB651SelfPumpWhileNoOpponentControlsCreatureOf(1, 1, "W", "U"),
		},
	})
}
