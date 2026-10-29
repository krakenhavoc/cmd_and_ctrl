package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dúnedain Blade — Artifact — Equipment {1}{W} (The Hobbit):
//
//	"Equipped creature gets +2/+1.
//	 Equip Human {1}
//	 Equip {3}"
//
// Two equip abilities, the Blackblade Reforged shape: the Human-only one
// is listed first with its tighter target clause and the cheaper cost,
// and the plain {3} one follows. Each is its own Spec.Activated entry, so
// either may be activated (CR 702.6d), and the restriction is checked
// while activating (CR 702.6c): a creature that stops being a Human after
// the cheap equip resolves keeps the Blade. The bonus is a plain layer-7c
// pump on whatever the Blade is attached to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b418846-6426-448a-b4cb-ce631e0a99a2",
		Name:         "Dúnedain Blade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
		},
		Activated: []ActivatedAbility{
			EquipOnlyAbility("Equip Human {1}", "{1}",
				TargetCreature("target Human you control", YouControl(), Subtype("Human"))),
			EquipAbility("{3}"),
		},
	})
}
