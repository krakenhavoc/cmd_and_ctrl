package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steppe Lynx — Creature — Cat {W}, 0/1 (Zendikar, #5583):
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
		OracleID:     "b16cef02-dcf8-4866-ac42-6ee929ddbea2",
		Name:         "Steppe Lynx",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Steppe Lynx — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Steppe Lynx — +2/+2 until end of turn")),
		},
	})
}
