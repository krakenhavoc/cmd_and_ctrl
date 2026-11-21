package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windrider Eel — Creature — Fish {3}{U}, 2/2 (Zendikar, #5583):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Flying is the printed keyword. Canopy Baloth's landfall pump: a land
// entering under its controller's control, the boost on the creature
// while it is still the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a3b16ae-210b-4d9b-83db-2801edfba394",
		Name:            "Windrider Eel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Windrider Eel — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Windrider Eel — +2/+2 until end of turn")),
		},
	})
}
