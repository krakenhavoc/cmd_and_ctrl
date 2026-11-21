package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f75bb13b-41fc-4614-b35e-f456069ce9c6",
		Name:         "Dwarven Mattock",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dwarven Mattock — attach it to target Dwarf you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target Dwarf you control", effects.Creature(), effects.HasSubtype("Dwarf")),
			),
			effects.WardAttached(effects.WardMana("{1}"), "Dwarven Mattock — ward {1}"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
