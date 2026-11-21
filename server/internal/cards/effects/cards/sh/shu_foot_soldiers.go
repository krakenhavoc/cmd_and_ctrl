package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu Foot Soldiers — Creature — Human Soldier {2}{W}, 2/3
// (Portal Three Kingdoms, #5622):
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
		OracleID:     "4d7ff9d5-dc06-4a70-a9c8-d6f88f02fd67",
		Name:         "Shu Foot Soldiers",
		Completeness: effects.CompletenessFull,
	})
}
