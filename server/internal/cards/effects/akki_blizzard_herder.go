package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akki Blizzard-Herder — Creature — Goblin Shaman {1}{R}, 1/1 (Betrayers
// of Kamigawa, #5602):
//
//	"When this creature dies, each player sacrifices a land of their
//	 choice."
//
// Hurloon Shaman's dies trigger: an edict over lands, each player
// (its controller included) choosing their own, in APNAP order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16fe1f04-9bbf-43a3-ab14-395fdf4a3a43",
		Name:         "Akki Blizzard-Herder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Akki Blizzard-Herder — each player sacrifices a land of their choice", Do(EachPlayerSacrifices{
				Match: Land(),
				Label: "a land",
			})),
		},
	})
}
