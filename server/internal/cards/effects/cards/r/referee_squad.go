package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Referee Squad — Creature — Homunculus {2}{U} 2/2 (The
// List): "Convoke. Vigilance. When this creature enters, tap target
// creature an opponent controls and put a stun counter on it."
//
// The tap-and-stun body is Floodpits Drowner's (tapAndStunTheTarget).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e14532e9-ecc2-4028-b99e-3629533962bc",
		Name:            "Referee Squad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		TapCost:         effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Referee Squad — tap target creature an opponent controls and put a stun counter on it", effects.TapAndStunTheTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
