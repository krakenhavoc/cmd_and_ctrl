package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Referee Squad — Creature — Homunculus {2}{U} 2/2 (The
// List): "Convoke. Vigilance. When this creature enters, tap target
// creature an opponent controls and put a stun counter on it."
//
// The tap-and-stun body is Floodpits Drowner's (tapAndStunTheTarget).
func init() {
	Register(Spec{
		OracleID:        "e14532e9-ecc2-4028-b99e-3629533962bc",
		Name:            "Referee Squad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		TapCost:         Convoke(),
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Referee Squad — tap target creature an opponent controls and put a stun counter on it", tapAndStunTheTarget),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
