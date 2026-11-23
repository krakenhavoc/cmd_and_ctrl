package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Second Sight — Instant {2}{U}:
//
//	"Choose one —
//	 • Look at the top five cards of target opponent's library, then put
//	   them back in any order.
//	 • Look at the top five cards of your library, then put them back in
//	   any order.
//	 Entwine {U} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). Each bullet is Portent's look and
// reorder (LookAtLibraryThenPlace, ADR 0088): only you see the cards,
// and you put them back on top in the order you choose.
// Entwined, the opponent's library is ordered first and your own after
// (CR 702.42b); the second look waits for the first answer (#2789).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1a4e948a-5a5e-4811-a6af-7ea93a051746",
		Name:          "Second Sight",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Entwine("{U}")},
		Modes: effects.Entwining(effects.ChooseOne(
			effects.ModeDoing("Look at the top five cards of target opponent's library, then put them back in any order.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.LookAtLibraryThenPlace{
						Owner:     t.ID,
						N:         5,
						Placement: game.LibraryPlaceTop,
						Label:     "Second Sight — put them back in any order",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Look at the top five cards of your library, then put them back in any order.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.LookAtLibraryThenPlace{
						N:         5,
						Placement: game.LibraryPlaceTop,
						Label:     "Second Sight — put them back in any order",
					}.Apply(ctx)
				}),
		)),
	})
}
