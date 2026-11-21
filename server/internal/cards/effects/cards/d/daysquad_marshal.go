package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daysquad Marshal — Creature — Human Soldier {3}{W}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, create a 1/1 white Human Soldier creature
//	 token."
//
// Honored Knight-Captain's entry trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42de2d46-c98d-4ae5-a53d-97531592529b",
		Name:         "Daysquad Marshal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Daysquad Marshal — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
