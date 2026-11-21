package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Automatic Librarian — Artifact Creature — Construct {3}, 3/2 (Dominaria United, S82):
//
//	"When this creature enters, scry 2."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb570331-35da-44b8-b711-ccf7444d8e39",
		Name:         "Automatic Librarian",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Automatic Librarian — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
