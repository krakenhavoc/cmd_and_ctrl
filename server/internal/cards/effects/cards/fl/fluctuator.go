package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fluctuator — Artifact {2} (Commander 2020, S82 OLD, #5513):
//
//	"Cycling abilities you activate cost {2} less to activate."
//
// Silver-Fur Master's activation discount over the cycling keyword bit
// (game.AbilityCostSubject.Cycling, set by Cycling and Typecycling, CR 702.29
// and 702.29e), so a cycling ability a static granted to a card in hand is
// matched too. It takes off generic mana only (CR 118.7a), so cycling {W} is
// not cheaper, and only the Fluctuator's controller's activations are cheaper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ed0c7cc-9969-4387-90f8-3084dd18026b",
		Name:         "Fluctuator",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsLess(2, "Cycling abilities you activate cost {2} less to activate.",
				effects.ACyclingAbilityCost(), effects.ActivatedByTheModifiersController()),
		},
	})
}
