package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warbringer — Creature — Orc Berserker {3}{R}, 3/3 (Dragons of Tarkir, S82 old
// batch 331):
//
//	"Dash costs you pay cost {2} less (as long as this creature is on the
//	 battlefield).
//	 Dash {2}{R} (You may cast this spell for its dash cost. If you do, it
//	 gains haste, and it's returned from the battlefield to its owner's hand at
//	 the beginning of the next end step.)"
//
// Dash is the shared alternative cost (CR 702.109). The reduction is Henzie's
// shape for blitz: a CostModifiers entry that reads whether the cast claims a
// dash cost (CostQuery.AltDash, a small seam), so it applies to every dash cost
// its controller pays, Warbringer's own included once another Warbringer is out,
// and never to the same creature cast for its mana cost. It comes off the generic
// part (CR 118.7a), so a {1}{R} dash cost becomes {R}, and never off a coloured
// symbol.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "0636090b-49e0-45b4-a68d-48b1cdc0c54d",
		Name:             "Warbringer",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Dash("{2}{R}")},
		CostModifiers:    []game.CostModifier{effects.DashCostsYouPayCostLess(2)},
	})
}
