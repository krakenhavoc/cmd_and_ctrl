package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrieking Mogg — Creature — Goblin {1}{R}, 1/1
// (Nemesis, #5618):
//
//	"Haste
//	 When this creature enters, tap all other creatures."
//
// Haste is printed data. The enters trigger is Timbermare's
// tapAllOtherCreatures: every creature but the Mogg, whoever controls
// it, as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "088a63d4-22d2-4061-8aef-2aa3fc348639",
		Name:            "Shrieking Mogg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shrieking Mogg — tap all other creatures", effects.TapAllOtherCreatures),
		},
	})
}
