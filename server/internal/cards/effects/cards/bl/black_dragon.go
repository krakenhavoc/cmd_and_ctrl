package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Dragon — Creature — Dragon {5}{B}{B}, 4/4 (Adventures in the Forgotten Realms):
//
//	Flying
//	 Acid Breath — When this creature enters, target creature an opponent controls gets -3/-3 until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad804712-18ba-49d7-9af2-4361edaa6cf8",
		Name:            "Black Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Black Dragon — Acid Breath: target creature an opponent controls gets -3/-3 until end of turn", effects.BoostTargetUntilEOT("Black Dragon — Acid Breath: -3/-3 until end of turn", -3, -3)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())), effects.ForTargets(effects.HarmUntilEOT(0, -3, -3))),
		},
	})
}
