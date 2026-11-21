package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alchemax Slayer-Bots —
//
// "When this creature enters, tap target creature an opponent controls and put
// a stun counter on it. (If a permanent with a stun counter would become
// untapped, remove one from it instead.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdc7356c-9621-4420-9fda-c8b91df0ec7c",
		Name:         "Alchemax Slayer-Bots",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Targeting(effects.WhenThisEnters("Alchemax Slayer-Bots — tap and stun target creature", effects.TapAndStunTheTarget), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))},
	})
}
