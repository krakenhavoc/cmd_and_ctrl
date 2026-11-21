package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wave-Wing Elemental — Creature — Elemental {5}{U}, 3/4 (Battle for
// Zendikar, #5547):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, this creature
//	 gets +2/+2 until end of turn."
//
// Canopy Baloth's landfall on a flyer. Landfall is the stack trigger on
// a land entering under its controller's control (CR 207.2c), so a land
// an opponent plays does nothing; the pump ends in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35044f8a-c8e0-41ae-ba61-43109e51f7e8",
		Name:            "Wave-Wing Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Wave-Wing Elemental — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Wave-Wing Elemental — +2/+2 until end of turn")),
		},
	})
}
