package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Leadfoot — Creature — Ogre {4}{R}, 3/3 (Mirrodin, #5606):
//
//	"Whenever this creature becomes blocked by an artifact creature, destroy
//	 that creature."
//
// Witherscale Wurm's per-pair block event (CR 509.3d: once for each creature
// that blocks the Ogre), read from the attacker's side, with the blocker
// required to be an artifact creature. The blocker is destroyed only while it
// is still on the battlefield as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17e00d6e-c57d-48ea-a1a1-0c1097083851",
		Name:         "Ogre Leadfoot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.OldB606BlockedByAnArtifactCreature,
				"Ogre Leadfoot — destroy that creature", effects.OldB606DestroyTheBlocker),
		},
	})
}
