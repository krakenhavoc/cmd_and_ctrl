package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Gatecrasher — Creature — Ogre Rogue {3}{R}, 3/3 (Dissension, #5598):
//
//	"When this creature enters, destroy target creature with defender."
//
// Bone Shredder's enters trigger with Smash to Dust's target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b88bc94-99c4-4be5-b474-a7dad68057b3",
		Name:         "Ogre Gatecrasher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ogre Gatecrasher — destroy target creature with defender", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with defender", effects.HasKeyword("defender"))),
		},
	})
}
