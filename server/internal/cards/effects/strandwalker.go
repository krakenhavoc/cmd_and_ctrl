package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strandwalker — Artifact — Equipment {5} (Mirrodin Besieged,
// #5576):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black
//	 Phyrexian Germ creature token, then attach this to it.)
//	 Equipped creature gets +2/+4 and has reach.
//	 Equip {4}"
//
// Batterskull's living weapon and equip shape (LivingWeapon, CR 702.92a)
// with a +2/+4 pump and a layer-6 reach grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55a4b186-c6d4-4b08-bf45-481429693948",
		Name:         "Strandwalker",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{LivingWeapon("Strandwalker")},
		Static: []game.StaticAbility{
			PumpAttached(2, 4),
			GrantToAttached("reach"),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
