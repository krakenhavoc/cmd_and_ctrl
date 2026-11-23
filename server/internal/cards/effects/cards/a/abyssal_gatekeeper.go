package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abyssal Gatekeeper — Creature — Horror {1}{B}, 1/1 (Weatherlight, #5629):
//
//	"When this creature dies, each player sacrifices a creature of their
//	 choice."
//
// Fleshbag Marauder's edict on a dies trigger: each player, its
// controller included, chooses a creature they control, and the chosen
// ones are sacrificed together (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4954181c-cb46-4cac-adc3-2c8b7693a211",
		Name:         "Abyssal Gatekeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisDies("Abyssal Gatekeeper — each player sacrifices a creature", effects.Do(effects.EachPlayerSacrifices{
				Match: effects.Creature(),
				Label: "a creature",
			}))),
		},
	})
}
