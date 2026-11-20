package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scheming Aspirant — Creature — Phyrexian Advisor {1}{B}, 1/3
// (Phyrexia: All Will Be One, #5653):
//
//	"Whenever you proliferate, each opponent loses 2 life and you gain 2
//	 life."
//
// WheneverYouProliferate: one trigger per proliferate its controller
// finishes, whatever was chosen (CR 701.34a), so "proliferate twice"
// drains twice. Life loss, not damage, and the gain is 2 however many
// opponents lost life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf4a3b1a-4b87-473e-be1f-3aa3da45c28e",
		Name:         "Scheming Aspirant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouProliferate("Scheming Aspirant — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
