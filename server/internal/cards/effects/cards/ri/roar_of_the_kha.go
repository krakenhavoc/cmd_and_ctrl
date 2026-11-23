package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roar of the Kha — Instant {1}{W}:
//
//	"Choose one —
//	 • Creatures you control get +1/+1 until end of turn.
//	 • Untap all creatures you control.
//	 Entwine {1}{W} (Choose both if you pay the entwine cost.)"
//
// Ramosian Rally's anthem (the creatures are fixed as it resolves, CR
// 611.2c) and Aggravated Assault's untap. Entwined (CR 702.42, #3866)
// the boost first, then the untap (CR 702.42b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9dddb69a-00a3-448b-8f8b-8e50cdd70d57",
		Name:          "Roar of the Kha",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{1}{W}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Creatures you control get +1/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Roar of the Kha"}.Apply(ctx)
				}),
			effects.ModeDoing("Untap all creatures you control.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.UntapAllCreaturesYouControl{}.Apply(ctx)
				}),
		)),
	})
}
