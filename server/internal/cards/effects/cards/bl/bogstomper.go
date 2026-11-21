package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bogstomper — Creature — Beast {4}{B}{B}, 6/5 (Core Set 2019, #5527):
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
		OracleID:     "7c2f1915-8d91-47c6-bab4-be325348673a",
		Name:         "Bogstomper",
		Completeness: effects.CompletenessFull,
	})
}
