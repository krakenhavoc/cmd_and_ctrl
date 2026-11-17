package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4954181c-cb46-4cac-adc3-2c8b7693a211",
		Name:         "Abyssal Gatekeeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisDies("Abyssal Gatekeeper — each player sacrifices a creature", Do(EachPlayerSacrifices{
				Match: Creature(),
				Label: "a creature",
			}))),
		},
	})
}
