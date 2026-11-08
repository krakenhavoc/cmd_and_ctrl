package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thinking Cap — Artifact — Equipment {1} (Murders at Karlov Manor,
// #4469):
//
//	"Equipped creature gets +1/+2.
//	 Equip Detective {1}
//	 Equip {3} (Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Dúnedain Blade's shape: the anthem on the equipped creature, a
// narrowed equip for a Detective you control (EquipOnlyAbility, so
// Leonin Shikari's permission still finds it) and the ordinary equip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a54fb3de-0581-461e-be33-dc4f4d16a33e",
		Name:         "Thinking Cap",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Detective {1}", "{1}",
				TargetCreature("target Detective you control", YouControl(), Subtype("Detective"))),
			EquipAbility("{3}"),
		},
	})
}
