package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brawler's Plate — Artifact — Equipment {3} (Magic 2015, #5556):
//
//	"Equipped creature gets +2/+2 and has trample.
//	 Equip {4}"
//
// Behemoth Sledge's two statics less the lifelink: a layer-7c pump and a
// layer-6 trample on the equipped creature, and the ordinary sorcery-speed
// equip row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9249adf2-8607-424f-a73b-fcddaf124af9",
		Name:         "Brawler's Plate",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("trample"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}
