package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blind Phantasm — Creature — Illusion {2}{U}, 2/3
// (Future Sight, #5594):
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
		OracleID:     "3b91be0e-0683-48df-91b0-25e90b492796",
		Name:         "Blind Phantasm",
		Completeness: effects.CompletenessFull,
	})
}
