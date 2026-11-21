package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shipwreck Patrol — Creature — Merfolk Warrior {3}{U}, 3/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"When this creature enters, tap target creature an opponent controls and
//	 put a stun counter on it."
//
// Cryotheory Adept's tap-and-stun body (tapTargetAndStunIt).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9851f207-9f75-4ceb-8024-8f9bf21f90ad",
		Name:         "Shipwreck Patrol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Shipwreck Patrol — tap target creature an opponent controls and put a stun counter on it", effects.TapTargetAndStunIt),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
