package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flameborn Viron — Creature — Phyrexian Insect {4}{R}{R}, 6/4 (New Phyrexia, #5575):
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
		OracleID:     "3fd2651d-89fb-4bd1-af4a-f7eae7ef9fd2",
		Name:         "Flameborn Viron",
		Completeness: effects.CompletenessFull,
	})
}
