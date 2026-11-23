package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hatching Plans — Enchantment {1}{U} (Enchanting Tales):
//
//	"When this enchantment is put into a graveyard from the battlefield, draw
//	 three cards."
//
// "Put into a graveyard from the battlefield" is what "dies" means for any
// permanent (CR 700.4), so the trigger is WhenThisDies; an exile or a bounce
// does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cebc9fcc-25ec-4366-a914-b95afbd134b0",
		Name:         "Hatching Plans",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Hatching Plans — draw three cards", effects.Do(effects.DrawCards{N: 3})),
		},
	})
}
