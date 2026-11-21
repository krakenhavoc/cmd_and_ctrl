package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colossapede — Creature — Insect {4}{G}, 5/5
// (Amonkhet, #5535):
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
		OracleID:     "b2cede1a-3a76-4658-bd12-e9df464d61e5",
		Name:         "Colossapede",
		Completeness: effects.CompletenessFull,
	})
}
