package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samurai's Katana — Artifact — Equipment {2}{R} (Final Fantasy, #4462):
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2, has trample and haste, and is a Samurai
//	 in addition to its other types.
//	 Murasame — Equip {5}"
//
// Bard's Bow's job-select shape: the shared JobSelect trigger, a layer 7c
// pump, a layer 6 keyword grant and a layer 4 creature-type addition. The
// Hero gets haste, so it can attack the turn the Katana comes down.
// "Murasame" is a flavor word with no rules meaning; the ability is an
// ordinary equip.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b8b5812-73bd-453e-bdff-aad1fafa9226",
		Name:         "Samurai's Katana",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("trample", "haste"),
			AttachedIsAlso("Samurai"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Samurai's Katana")},
		Activated: []ActivatedAbility{EquipAbility("{5}")},
	})
}
