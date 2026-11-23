package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trokin High Guard — Creature — Human Knight {3}{W}, 3/3 (Portal Second Age,
// #5625):
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
		OracleID:     "51af8662-07e6-4771-8ff2-c54364987698",
		Name:         "Trokin High Guard",
		Completeness: effects.CompletenessFull,
	})
}
