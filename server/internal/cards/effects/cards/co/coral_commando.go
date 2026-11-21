package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coral Commando — Creature — Merfolk Warrior {2}{U}, 3/2 (Ravnica
// Allegiance, #5521):
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
		OracleID:     "03ce2235-50d9-4f67-b588-166b9b37c664",
		Name:         "Coral Commando",
		Completeness: effects.CompletenessFull,
	})
}
