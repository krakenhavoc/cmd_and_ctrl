package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Colossodon Yearling — Creature — Beast {2}{G}, 2/4 (Dragons of Tarkir,
// #5551):
//
//	"(no rules text)"
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a322ad80-e64b-4e9a-b542-3a036dc14d28",
		Name:         "Colossodon Yearling",
		Completeness: effects.CompletenessFull,
	})
}
