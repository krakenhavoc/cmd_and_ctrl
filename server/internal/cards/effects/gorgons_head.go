package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "190cdc23-fa05-475f-b853-471ca0f21b71",
		Name:         "Gorgon's Head",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("deathtouch"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
