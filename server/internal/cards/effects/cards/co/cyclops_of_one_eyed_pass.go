package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cyclops of One-Eyed Pass — Creature — Cyclops {2}{R}{R}, 5/2 (Born of the Gods,
// #5559):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78922912-d7b9-47cc-8908-f44d7cc3fc4b",
		Name:         "Cyclops of One-Eyed Pass",
		Completeness: effects.CompletenessFull,
	})
}
