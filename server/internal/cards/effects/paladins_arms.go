package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paladin's Arms — Artifact — Equipment {2}{W} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+1, has ward {1}, and is a Knight in
//	 addition to its other types.
//	 Lightbringer and Hero's Shield — Equip {4}"
//
// The job-select shape (Bard's Bow) with Dwarven Mattock's granted ward:
// ward is a triggered ability (CR 702.21a), so the Equipment carries it
// as WardAttached rather than a layer-6 keyword. "Lightbringer and
// Hero's Shield" is a flavor word with no rules meaning; the ability is
// an ordinary equip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22076484-e9d5-4f33-8cf3-b9b427436b05",
		Name:         "Paladin's Arms",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			AttachedIsAlso("Knight"),
		},
		Triggered: []game.TriggeredAbility{
			JobSelect("Paladin's Arms"),
			WardAttached(WardMana("{1}"), "Paladin's Arms — ward {1}"),
		},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
