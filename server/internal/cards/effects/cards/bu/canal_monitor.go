package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canal Monitor — Creature — Lizard {4}{B}, 5/3 (Rivals of Ixalan, #5531):
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
		OracleID:     "9dadeecc-1d4b-4799-9ac9-4ca492006e46",
		Name:         "Canal Monitor",
		Completeness: effects.CompletenessFull,
	})
}
