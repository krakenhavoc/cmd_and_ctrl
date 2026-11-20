package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbed Battlegear — Artifact — Equipment {3} (Scars of Mirrodin, #5577):
//
//	"Equipped creature gets +4/-1.
//	 Equip {2}"
//
// Leonin Scimitar's shape: PumpAttached and the equip row. A 1-toughness
// creature equipped with it dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b81fed36-07c4-4903-b5b4-24e62d2ce239",
		Name:         "Barbed Battlegear",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(4, -1)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
