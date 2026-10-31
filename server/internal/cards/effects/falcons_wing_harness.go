package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falcon's Wing Harness — {1}{U} Artifact — Equipment:
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+1 and has flying and ward {1}.
//	 Equip {2}{U}"
//
// The entry trigger is Maul of the Skyclaves' shape; the ward is granted to
// the equipped creature the way Armguard Familiar grants its own.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "e69db5f0-beed-48e8-bc25-768b1e8be5ea",
		Name:         "Falcon's Wing Harness",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Falcon's Wing Harness — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
			WardAttached(WardMana("{1}"), "Falcon's Wing Harness — equipped creature's ward {1}"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}{U}"),
		},
	})
}
