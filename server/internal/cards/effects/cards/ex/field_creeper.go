package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field Creeper — Artifact Creature — Scarecrow {2}, 2/1 (Eldritch Moon,
// #5541):
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
		OracleID:     "e6584e39-099c-437f-9239-fa718cac85af",
		Name:         "Field Creeper",
		Completeness: effects.CompletenessFull,
	})
}
