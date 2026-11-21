package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piggy Bank — Artifact Creature — Boar Toy {1}{R}, 3/2 (Duskmourn:
// House of Horror, #4466):
//
//	"When this creature dies, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Common Crook's death trigger on a red 3/2: the Treasure is created
// under the dead creature's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a13f4095-300b-434f-aa9a-3608aaae5163",
		Name:         "Piggy Bank",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Piggy Bank — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
