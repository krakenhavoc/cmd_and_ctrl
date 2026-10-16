package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "9dddb69a-00a3-448b-8f8b-8e50cdd70d57",
		Name:          "Roar of the Kha",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{1}{W}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Creatures you control get +1/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Roar of the Kha"}.Apply(ctx)
				}),
			ModeDoing("Untap all creatures you control.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return UntapAllCreaturesYouControl{}.Apply(ctx)
				}),
		)),
	})
}
