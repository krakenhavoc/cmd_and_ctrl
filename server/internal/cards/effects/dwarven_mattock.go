package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Mattock — Artifact — Equipment {2}:
//
//	"When this Equipment enters, attach it to target Dwarf you control.
//	 Equipped creature gets +2/+2 and has ward {1}.
//	 Equip {3}"
//
// The attach is Embercleave's and Inventor's Axe's: a targeted enters
// trigger resolving through AttachSourceToTarget, with the target
// narrowed to a Dwarf. With no Dwarf to name the trigger is removed
// from the stack and the Mattock simply stays unattached (CR 603.3d).
// The ward is Lavaspur Boots' WardAttached, the Equipment carrying the
// trigger that watches its host.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f75bb13b-41fc-4614-b35e-f456069ce9c6",
		Name:         "Dwarven Mattock",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dwarven Mattock — attach it to target Dwarf you control", AttachSourceToTarget),
				PermanentYouControl("target Dwarf you control", Creature(), HasSubtype("Dwarf")),
			),
			WardAttached(WardMana("{1}"), "Dwarven Mattock — ward {1}"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
