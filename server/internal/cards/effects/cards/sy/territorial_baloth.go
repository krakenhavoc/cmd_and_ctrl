package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Baloth — Creature — Beast {4}{G}, 4/4 (Zendikar, #5583):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Canopy Baloth's landfall pump: a land entering under its controller's
// control, the boost on the creature while it is still the same object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f772e767-a931-4891-94eb-d92a8ca7f8c1",
		Name:         "Territorial Baloth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Territorial Baloth — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Territorial Baloth — +2/+2 until end of turn")),
		},
	})
}
