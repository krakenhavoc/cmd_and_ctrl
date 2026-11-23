package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeletal Snake — Creature — Snake Skeleton {1}{B}, 2/1 (Portal, #5630):
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
		OracleID:     "7af786a0-0851-4f87-ad0d-36b40527700c",
		Name:         "Skeletal Snake",
		Completeness: effects.CompletenessFull,
	})
}
