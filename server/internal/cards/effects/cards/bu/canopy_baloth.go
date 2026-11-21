package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canopy Baloth — Creature — Beast {3}{G}, 4/3 (Foundations Jumpstart):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Boughside Wanderers' landfall pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4493828f-144f-49cf-9fd2-61de86e632ec",
		Name:         "Canopy Baloth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Canopy Baloth — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Canopy Baloth — +2/+2 until end of turn")),
		},
	})
}
