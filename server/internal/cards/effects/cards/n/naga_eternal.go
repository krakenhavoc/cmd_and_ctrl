package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naga Eternal — Creature — Zombie Snake {2}{U}, 3/2 (War of the Spark, #5520):
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
		OracleID:     "9d6a10c6-5d39-49c6-b50d-8fc5b8ddcd77",
		Name:         "Naga Eternal",
		Completeness: effects.CompletenessFull,
	})
}
