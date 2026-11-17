package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hurloon Shaman — Creature — Minotaur Shaman {1}{R}{R}, 2/3 (Weatherlight, #5629):
//
//	"When this creature dies, each player sacrifices a land of their
//	 choice."
//
// The edict over lands: each player, its controller included, chooses a
// land they control, and the chosen ones are sacrificed together
// (CR 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fc74fc9-bb02-4d9e-bb8e-f44f219d8946",
		Name:         "Hurloon Shaman",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hurloon Shaman — each player sacrifices a land", Do(EachPlayerSacrifices{
				Match: Land(),
				Label: "a land",
			})),
		},
	})
}
