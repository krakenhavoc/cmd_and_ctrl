package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Dragon — Creature — Dragon {5}{B}{B}, 4/4 (Adventures in the Forgotten Realms):
//
//	Flying
//	 Acid Breath — When this creature enters, target creature an opponent controls gets -3/-3 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad804712-18ba-49d7-9af2-4361edaa6cf8",
		Name:            "Black Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Black Dragon — Acid Breath: target creature an opponent controls gets -3/-3 until end of turn", boostTargetUntilEOT("Black Dragon — Acid Breath: -3/-3 until end of turn", -3, -3)),
				TargetCreature("target creature an opponent controls", OpponentControls())), ForTargets(HarmUntilEOT(0, -3, -3))),
		},
	})
}
