package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alchemax Slayer-Bots —
//
// "When this creature enters, tap target creature an opponent controls and put
// a stun counter on it. (If a permanent with a stun counter would become
// untapped, remove one from it instead.)"
func init() {
	Register(Spec{
		OracleID:     "cdc7356c-9621-4420-9fda-c8b91df0ec7c",
		Name:         "Alchemax Slayer-Bots",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Targeting(WhenThisEnters("Alchemax Slayer-Bots — tap and stun target creature", tapAndStunTheTarget), TargetCreature("target creature an opponent controls", OpponentControls()))},
	})
}
