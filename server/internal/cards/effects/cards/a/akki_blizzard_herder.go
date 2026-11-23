package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "16fe1f04-9bbf-43a3-ab14-395fdf4a3a43",
		Name:         "Akki Blizzard-Herder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Akki Blizzard-Herder — each player sacrifices a land of their choice", effects.Do(effects.EachPlayerSacrifices{
				Match: effects.Land(),
				Label: "a land",
			})),
		},
	})
}
