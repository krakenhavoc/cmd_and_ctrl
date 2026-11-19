package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hookblade — Artifact — Equipment {1}{W} (Assassin's Creed, #5948):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +1/+0.
//	 During your turn, equipped creature has flying.
//	 Equip {2}"
//
// Baseball Bat's entry attach, the pump, and Bilbo's Ring's layer 6 grant that
// holds only while it is the Equipment controller's turn ("your" is the
// Hookblade's controller, not the creature's). The flying comes and goes with
// the turn without any bookkeeping.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8df6c4e2-8b62-4c64-a1d2-f17ae9d02bac",
		Name:         "Hookblade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttachedWhile(duringSourceControllersTurn, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Hookblade — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
