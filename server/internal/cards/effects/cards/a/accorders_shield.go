package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accorder's Shield — Artifact — Equipment {0} (Scars of Mirrodin, #5577):
//
//	"Equipped creature gets +0/+3 and has vigilance.
//	 Equip {3}"
//
// Loxodon Warhammer's two statics with other numbers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd70d439-bd60-43d1-ac26-13b744cc4a37",
		Name:         "Accorder's Shield",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 3),
			effects.GrantToAttached("vigilance"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
