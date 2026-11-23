package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonework Puma — Artifact Creature — Cat Ally {3}, 2/2 (Zendikar, #5583):
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
		OracleID:     "841b1010-a7d3-415e-b76d-7b50fea909cb",
		Name:         "Stonework Puma",
		Completeness: effects.CompletenessFull,
	})
}
