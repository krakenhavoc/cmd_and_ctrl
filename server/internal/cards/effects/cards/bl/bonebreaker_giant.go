package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonebreaker Giant — Creature — Giant {4}{R}, 4/4 (Magic 2012, #5573):
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
		OracleID:     "69166587-3db4-4eb0-b153-b75021be9a36",
		Name:         "Bonebreaker Giant",
		Completeness: effects.CompletenessFull,
	})
}
