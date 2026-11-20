package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vulshok Morningstar — Artifact — Equipment {2} (Darksteel, #5605):
//
//	"Equipped creature gets +2/+2.
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as
//	 a sorcery.)"
//
// Leonin Scimitar's shape: the shared attached pump and equip ability
// (CR 702.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12a8adc4-927f-4314-b2ef-9c647ace68d5",
		Name:         "Vulshok Morningstar",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
