package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Type-2 Phaser — Artifact — Equipment {2} (Star Trek, #4973):
//
//	"When this Equipment enters, tap target creature an opponent controls
//	 and put a stun counter on it.
//	 Equipped creature gets +3/+0.
//	 Equip {3}"
//
// Enterprise NX-01's tap-and-stun enters trigger (tapTargetAndStunIt) on
// an Equipment: PumpAttached(3, 0) and EquipAbility("{3}"). The trigger
// targets as it goes on the stack and does nothing to a creature that is
// no longer a legal target as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d4d3682-e904-412d-b8ae-734962865dfb",
		Name:         "Type-2 Phaser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Type-2 Phaser — tap target creature an opponent controls and put a stun counter on it", effects.TapTargetAndStunIt),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Static:    []game.StaticAbility{effects.PumpAttached(3, 0)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
