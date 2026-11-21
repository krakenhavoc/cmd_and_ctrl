package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meditation Puzzle — Instant {3}{W}{W} (Magic 2015, #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 You gain 8 life."
//
// Spring of Eternal Peace's life gain at instant speed, with the convoke
// TapCost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "724c29a2-1783-401e-887c-2189afd5a944",
		Name:         "Meditation Puzzle",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 8}),
	})
}
