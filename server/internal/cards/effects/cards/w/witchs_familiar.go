package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch's Familiar — Creature — Frog {2}{B}, 2/3 (Magic 2015, #5556):
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
		OracleID:     "abe06b4d-e7f0-4125-8cfc-e36296e8bacd",
		Name:         "Witch's Familiar",
		Completeness: effects.CompletenessFull,
	})
}
