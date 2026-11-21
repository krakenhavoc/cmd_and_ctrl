package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dutiful Servants — Creature — Zombie {3}{W}, 2/5 (Hour of Devastation, #5534):
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
		OracleID:     "5837b3d9-b3cb-4333-b0db-ed38e49c997f",
		Name:         "Dutiful Servants",
		Completeness: effects.CompletenessFull,
	})
}
