package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Reaper — Creature — Phyrexian Zombie {4}{B}, 3/3 (Invasion, S82 older sets batch 664, #5616):
//
//	"Whenever this creature becomes blocked by a green creature, destroy that
//	 creature. It can't be regenerated."
//
// Ogre Leadfoot's per-blocker block trigger (CR 509.3d: one for each creature
// that blocks the Reaper) with a green blocker, and a destroy that ignores
// regeneration (CR 701.19c). The blocker is destroyed only while it is still on
// the battlefield as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "245c3681-fc47-47fe-b6e5-34c96315c135",
		Name:         "Phyrexian Reaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.OldB664BlockedByAGreenCreature,
				"Phyrexian Reaper — destroy that creature; it can't be regenerated", effects.OldB664DestroyTheBlockerNoRegen),
		},
	})
}
