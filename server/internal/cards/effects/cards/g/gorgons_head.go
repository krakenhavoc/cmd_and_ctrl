package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorgon's Head — Artifact — Equipment {1} (Born of the Gods,
// #5559):
//
//	"Equipped creature has deathtouch.
//	 Equip {2}"
//
// Basilisk Collar with deathtouch alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "190cdc23-fa05-475f-b853-471ca0f21b71",
		Name:         "Gorgon's Head",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("deathtouch"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
