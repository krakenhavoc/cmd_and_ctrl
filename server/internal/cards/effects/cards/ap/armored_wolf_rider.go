package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armored Wolf-Rider — Creature — Elf Knight {3}{G}{W}, 4/6:
//
//	(no rules text)
//
// A vanilla creature: the printed 4/6 body is the whole card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0965cd1-18fa-4ccc-8c5f-1414c6567d67",
		Name:         "Armored Wolf-Rider",
		Completeness: effects.CompletenessFull,
	})
}
