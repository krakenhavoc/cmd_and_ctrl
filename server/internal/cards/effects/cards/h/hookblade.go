package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8df6c4e2-8b62-4c64-a1d2-f17ae9d02bac",
		Name:         "Hookblade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttachedWhile(effects.DuringSourceControllersTurn, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Hookblade — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
