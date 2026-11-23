package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotted Hystrix — Creature — Phyrexian Beast {4}{G}, 3/6 (New Phyrexia, #5575):
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
		OracleID:     "e600afe3-8d5b-46ab-b98c-f0bacb58ab2a",
		Name:         "Rotted Hystrix",
		Completeness: effects.CompletenessFull,
	})
}
