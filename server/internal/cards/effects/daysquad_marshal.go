package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daysquad Marshal — Creature — Human Soldier {3}{W}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, create a 1/1 white Human Soldier creature
//	 token."
//
// Honored Knight-Captain's entry trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42de2d46-c98d-4ae5-a53d-97531592529b",
		Name:         "Daysquad Marshal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Daysquad Marshal — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
