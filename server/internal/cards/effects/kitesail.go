package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitesail — Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+0 and has flying.
//	 Equip {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b079f9db-974d-4525-a894-57b754ba9dcc",
		Name:         "Kitesail",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("flying"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
