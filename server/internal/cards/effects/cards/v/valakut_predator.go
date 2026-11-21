package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valakut Predator — Creature — Elemental {2}{R}, 2/2 (Battle for
// Zendikar, #5547):
//
//	"Landfall — Whenever a land you control enters, this creature
//	 gets +2/+2 until end of turn."
//
// Canopy Baloth's landfall. Landfall is the stack trigger on a land
// entering under its controller's control (CR 207.2c), so a land an
// opponent plays does nothing; the pump ends in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b86a5d90-0820-46cb-b43c-c03a423a3a9f",
		Name:         "Valakut Predator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Valakut Predator — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Valakut Predator — +2/+2 until end of turn")),
		},
	})
}
