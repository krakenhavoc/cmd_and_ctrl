package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valiant Guard — Creature — Human Soldier {W}, 0/3 (Conflux, #5587):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it. The cost, types
// and body ride game.Card from Scryfall, so the Spec is the key and the
// name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b00b4c2-d5c4-4cf7-bdcf-cd96297c4350",
		Name:         "Valiant Guard",
		Completeness: effects.CompletenessFull,
	})
}
