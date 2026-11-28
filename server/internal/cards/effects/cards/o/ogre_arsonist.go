package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ogre Arsonist — Creature — Ogre {4}{R}, 3/3 (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"When this creature enters, destroy target land."
//
// The land is chosen as the trigger goes on the stack; one that left is skipped on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00af96af-5eae-4044-a2f7-a08cd0699d1e",
		Name:         "Ogre Arsonist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ogre Arsonist — destroy target land", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
