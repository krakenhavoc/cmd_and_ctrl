package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Axebane Stag — Creature — Elk {6}{G}, 6/7 (Return to Ravnica, #5566):
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
		OracleID:     "a45d7135-c042-40d9-818c-74fa1c7af8ad",
		Name:         "Axebane Stag",
		Completeness: effects.CompletenessFull,
	})
}
