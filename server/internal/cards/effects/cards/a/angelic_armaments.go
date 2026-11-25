package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Armaments — Artifact — Equipment {3} (Commander Legends):
//
//	"Equipped creature gets +2/+2, has flying, and is a white Angel in
//	 addition to its other colors and types.
//	 Equip {4}"
//
// The Angel type (layer 4) and the colour (layer 5) are additions, not
// replacements.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2042b38-1f06-44d4-b56f-2cb4d9197712",
		Name:         "Angelic Armaments",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
			effects.AttachedIsAlso("Angel"),
			effects.OldB084AttachedIsColorInAddition("W", "Equipped creature is white in addition to its other colors."),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
