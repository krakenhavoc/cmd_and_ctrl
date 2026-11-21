package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenging Hunter — Creature — Dragon Ranger {4}{G}, 5/4:
//
//	"Trample
//	 When this creature enters, you take the initiative."
//
// Trample is engine-side; the ETB is TakeTheInitiative (ADR 0203,
// #4096).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0aa5e3d-17c0-4ac7-9d76-9ce579d1eb2b",
		Name:            "Avenging Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Avenging Hunter"),
		},
	})
}
