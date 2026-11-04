package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skateboard — {1} Artifact — Equipment (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"When this Equipment enters, tap target permanent.
//	 Equipped creature gets +1/+0 and has haste.
//	 Equip {1}"
//
// Lavaspur Boots's statics and equip (CR 702.6a), plus an enters trigger
// that taps its target (tapTheFirstLegalTarget, Storm Elemental's body).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "decfde5e-ea53-4b15-8f69-51da7ef39d15",
		Name:         "Skateboard",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Skateboard — tap target permanent", tapTheFirstLegalTarget),
				TargetPermanent("target permanent")),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
