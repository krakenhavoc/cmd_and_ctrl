package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shatterskull Giant — Creature — Giant Warrior {2}{R}{R}, 4/3 (Zendikar, #5583):
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
		OracleID:     "aa241f75-ce34-49a8-8ffd-785172249bcf",
		Name:         "Shatterskull Giant",
		Completeness: effects.CompletenessFull,
	})
}
