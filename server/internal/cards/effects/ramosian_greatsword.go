package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ramosian Greatsword — Artifact — Equipment {4}{R} (March of the
// Machine): "Convoke. Equipped creature gets +3/+1 and has trample.
// Equip {2}"
func init() {
	Register(Spec{
		OracleID:     "8c53a177-c8c8-4b9f-a221-f205da67daa5",
		Name:         "Ramosian Greatsword",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Static: []game.StaticAbility{
			PumpAttached(3, 1),
			GrantToAttached("trample"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
