package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jackhammer — Artifact — Equipment {1}{R} (Streets of New Capenna, #5491):
//
//	"Equipped creature gets +2/+0.
//	 Equip {2}"
//
// A plain pump Equipment: PumpAttached and the equip ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f73d89ec-05c9-4998-a24b-bc675b04ec16",
		Name:         "Jackhammer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
