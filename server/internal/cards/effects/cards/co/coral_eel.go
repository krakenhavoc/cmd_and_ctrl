package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coral Eel — Creature — Fish {1}{U}, 2/1 (Portal, #5630):
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
		OracleID:     "10706fd1-7847-4316-be8d-59b56143ce45",
		Name:         "Coral Eel",
		Completeness: effects.CompletenessFull,
	})
}
