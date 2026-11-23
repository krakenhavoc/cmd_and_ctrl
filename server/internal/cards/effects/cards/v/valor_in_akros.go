package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valor in Akros — Enchantment {3}{W} (Jumpstart 2022, #5948):
//
//	"Whenever a creature you control enters, creatures you control get +1/+1
//	 until end of turn."
//
// Ampryn Tactician's mass pump behind an entry trigger: the creatures you
// control as the trigger resolves, the one that entered included, are fixed
// at resolution (CR 611.2c), so a creature that enters later is not pumped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56a60fcb-f9b7-438f-9221-eaca8fb2649b",
		Name:         "Valor in Akros",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Valor in Akros — creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 1, Toughness: 1,
					Label: "Valor in Akros — +1/+1",
				})),
		},
	})
}
