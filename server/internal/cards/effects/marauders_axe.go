package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marauder's Axe — Artifact — Equipment {2} (Core Set 2019, #5527):
//
//	"Equipped creature gets +2/+0.
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)"
//
// Vanquisher's Axe's text: the equipped creature's +2/+0 and equip {2}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d0e49fa-5dfb-48c2-af97-fdfb788d5f40",
		Name:         "Marauder's Axe",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 0)},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
