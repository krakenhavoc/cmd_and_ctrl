package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wetland Sambar — Creature — Elk {1}{U}, 2/1 (Khans of Tarkir, #5554):
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
		OracleID:     "8a50b8f7-1af8-4e00-9734-66a87e05cc6f",
		Name:         "Wetland Sambar",
		Completeness: effects.CompletenessFull,
	})
}
