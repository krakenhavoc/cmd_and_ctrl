package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cebc9fcc-25ec-4366-a914-b95afbd134b0",
		Name:         "Hatching Plans",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hatching Plans — draw three cards", Do(DrawCards{N: 3})),
		},
	})
}
