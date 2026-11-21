package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cobblebrute — Creature — Elemental {3}{R}, 5/2 (Return to Ravnica, #5566):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d4ed1f2-0918-444c-95ef-d58beb5aa216",
		Name:         "Cobblebrute",
		Completeness: effects.CompletenessFull,
	})
}
