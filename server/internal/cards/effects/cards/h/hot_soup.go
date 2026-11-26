package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hot Soup — Artifact — Equipment {1} (Magic 2015, #5556):
//
//	"Equipped creature can't be blocked.
//	 Whenever equipped creature is dealt damage, destroy it.
//	 Equip {3}"
//
// Whispersilk Cloak's unblockable restriction and Mire Blight's damage
// trigger, read through the Equipment's attachment. The creature is named
// with its object epoch when the damage is dealt, and destroyed only if it is
// still that creature on the battlefield (CR 603.7c, 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a853e48b-1e15-4492-a5b4-9924b1a4cf9a",
		Name:         "Hot Soup",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantBeBlocked)},
		Triggered:    []game.TriggeredAbility{effects.DestroyItWhenAttachedCreatureIsDealtDamage("Hot Soup — destroy equipped creature")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
