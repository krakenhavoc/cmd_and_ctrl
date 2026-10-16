package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Journey of Discovery — Sorcery {2}{G}:
//
//	"Choose one —
//	 • Search your library for up to two basic land cards, reveal them,
//	   put them into your hand, then shuffle.
//	 • You may play up to two additional lands this turn.
//	 Entwine {2}{G} (Choose both if you pay the entwine cost.)"
//
// The first bullet is Burnished Hart's search with the hand as the
// destination; the second is Explore's grant, two land plays. Entwined
// (CR 702.42, #3866) the search resolves first (CR 702.42b) and the
// grant waits for it (the walk holds the later bullet behind the search
// prompt, #2789).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "1c586d8a-9d1a-48a7-bb3e-9b2c0c329f8d",
		Name:          "Journey of Discovery",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}{G}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Search your library for up to two basic land cards, reveal them, put them into your hand, then shuffle.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     2,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Journey of Discovery — up to two basic land cards",
					}.Apply(ctx)
				}),
			ModeDoing("You may play up to two additional lands this turn.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 2)
					return nil
				}),
		)),
	})
}
