package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tooth and Nail — Sorcery {5}{G}{G}:
//
//	"Choose one —
//	 • Search your library for up to two creature cards, reveal them,
//	   put them into your hand, then shuffle.
//	 • Put up to two creature cards from your hand onto the battlefield.
//	 Entwine {2} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). The search is SearchLibrary with a
// stated quality, so finding fewer than two (or none) is an answer (CR
// 701.23b); the put is PutFromHandOntoBattlefield, up to two creature
// cards entering as one batch. Entwined, the search finishes first and
// the creatures it found are in hand to be put onto the battlefield (CR
// 702.42b; the second bullet waits for the search answer, #2789).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "382b7873-e78d-473b-8346-d92d16e8f634",
		Name:          "Tooth and Nail",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Search your library for up to two creature cards, reveal them, put them into your hand, then shuffle.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return c.IsCreature() },
						Dest:      game.ZoneHand,
						Limit:     2,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Tooth and Nail — search for up to two creature cards",
						Source:    ctx.Source(),
					}.Apply(ctx)
				}),
			ModeDoing("Put up to two creature cards from your hand onto the battlefield.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return PutFromHandOntoBattlefield{
						Match:    Creature(),
						Optional: true,
						Max:      2,
						Label:    "Tooth and Nail — put up to two creature cards from your hand onto the battlefield",
					}.Apply(ctx)
				}),
		)),
	})
}
