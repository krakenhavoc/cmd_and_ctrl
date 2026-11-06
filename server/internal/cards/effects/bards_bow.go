package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bard's Bow — Artifact — Equipment {2}{G} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2, has reach, and is a Bard in addition
//	 to its other types.
//	 Perseus's Bow — Equip {6}"
//
// The job-select Equipment shape (Sage's Nouliths, Dancer's Chakrams):
// the shared JobSelect trigger, a layer 7c pump, a layer 6 keyword grant
// and a layer 4 creature-type addition. "Perseus's Bow" is a flavor word
// with no rules meaning; the ability is an ordinary equip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4aceb786-960d-4719-9de8-79dffd491a51",
		Name:         "Bard's Bow",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("reach"),
			AttachedIsAlso("Bard"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Bard's Bow")},
		Activated: []ActivatedAbility{EquipAbility("{6}")},
	})
}
