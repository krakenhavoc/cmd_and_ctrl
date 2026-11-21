package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underdark Explorer — Creature — Lizard Warrior {4}{B}, 5/3:
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 When this creature enters, you take the initiative."
//
// Menace is engine-side; the ETB is TakeTheInitiative (ADR 0203,
// #4096).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7647682d-d491-4cb1-add7-b4d745994e62",
		Name:            "Underdark Explorer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Underdark Explorer"),
		},
	})
}
