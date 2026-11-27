package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Axegrinder Giant — Creature — Giant Warrior {4}{R}{R}, 6/4 (Lorwyn, S82 OLD,
// #5593):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go). The
// cost, types and body ride game.Card from Scryfall, so the Spec is the key and
// the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5355c5e-d07d-48ca-90c5-98edcab1a3c7",
		Name:         "Axegrinder Giant",
		Completeness: effects.CompletenessFull,
	})
}
