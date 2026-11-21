package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasping Longneck — Enchantment Creature — Horror {2}{G}, 4/2
// (Duskmourn: House of Horror, #4466):
//
//	"Reach
//	 When this creature dies, you gain 2 life."
//
// Reach rides PrintedKeywords. The death trigger is Piggy Bank's with a
// life gain for the dead creature's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc509d30-ceca-47ef-a574-62c021812014",
		Name:            "Grasping Longneck",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Grasping Longneck — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
