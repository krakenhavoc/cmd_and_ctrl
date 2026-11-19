package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cliffhaven Kitesail — Artifact — Equipment {1} (Foundations Jumpstart):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature has flying.
//	 Equip {2}"
//
// Barbed Bloodletter's entry attach with a flying grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68fb8f21-4403-4829-8826-f2418664197e",
		Name:         "Cliffhaven Kitesail",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Cliffhaven Kitesail — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
