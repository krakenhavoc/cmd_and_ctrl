package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Accorder's Shield — Artifact — Equipment {0} (Scars of Mirrodin, #5577):
//
//	"Equipped creature gets +0/+3 and has vigilance.
//	 Equip {3}"
//
// Loxodon Warhammer's two statics with other numbers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd70d439-bd60-43d1-ac26-13b744cc4a37",
		Name:         "Accorder's Shield",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(0, 3),
			GrantToAttached("vigilance"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
