package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pennon Blade — Artifact — Equipment {3} (Rise of the Eldrazi, #5581):
//
//	"Equipped creature gets +1/+1 for each creature you control.
//	 Equip {4} ({4}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Kemba's Banner's count: "you" is the Equipment's controller, and the
// equipped creature counts itself when that player controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b97e6acd-a676-429e-b875-b734759ebea7",
		Name:         "Pennon Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttachedPer(1, 1, roeB1CreaturesYouControl)},
		Activated:    []ActivatedAbility{EquipAbility("{4}")},
	})
}
