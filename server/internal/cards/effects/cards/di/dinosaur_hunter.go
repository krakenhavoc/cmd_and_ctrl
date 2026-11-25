package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dinosaur Hunter — Creature — Human Pirate {1}{B}, 2/2 (Rivals of Ixalan, S82
// OLD batch 217, #5531):
//
//	"Whenever this creature deals damage to a Dinosaur, destroy that creature."
//
// Spider-Slayer's trigger for any damage (not only combat damage) to a
// Dinosaur, as the shared constructor builds it: the damaged creature is named
// as the object it was when the trigger fired, so one that died to the damage
// is not destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c58f2eb-2617-4d10-a393-6e4a9eddd1dd",
		Name:         "Dinosaur Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDamagesACreatureOfTypeDestroyIt("Dinosaur Hunter — destroy the Dinosaur it dealt damage to", "Dinosaur"),
		},
	})
}
