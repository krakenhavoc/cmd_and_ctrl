package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Machinist's Arsenal — Artifact — Equipment {4}{W} (Final Fantasy,
// #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2 for each artifact you control and is an
//	 Artificer in addition to its other types.
//	 Machina — Equip {4}"
//
// The job-select shape (Bard's Bow) with Cranial Plating's counted pump:
// "you" is the Equipment's controller, and the count is read on every
// recompute, so it tracks artifacts coming and going. The Arsenal is an
// artifact itself, so it always counts at least one. "Machina" is a
// flavor word with no rules meaning; the ability is an ordinary equip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7fff34e-9b80-4a67-a917-fd9efddb5a1f",
		Name:         "Machinist's Arsenal",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(2, 2, artifactsControlledBy),
			AttachedIsAlso("Artificer"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Machinist's Arsenal")},
		Activated: []ActivatedAbility{EquipAbility("{4}")},
	})
}
