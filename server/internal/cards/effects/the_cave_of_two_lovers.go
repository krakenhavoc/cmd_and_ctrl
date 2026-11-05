package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Cave of Two Lovers — Enchantment — Saga {3}{R}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create two 1/1 white Ally creature tokens.
//	 II — Search your library for a Mountain or Cave card, reveal it,
//	      put it into your hand, then shuffle.
//	 III — Earthbend 3."
//
// Chapter I is the set's 1/1 white Ally token row, two of them.
// Chapter II is There and Back Again's land-type search sent to the hand
// and revealed: "Mountain" and "Cave" are subtype words read through
// the Saga's text (CR 612.2), any card with either subtype qualifies,
// and the searcher may find nothing (CR 701.23b). Chapter III is the
// earthbend keyword action on a target land its controller controls; a
// land that has become an illegal target makes it do nothing (the
// ruling). The Saga is sacrificed after III by the engine (CR 714.4).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	const name = "The Cave of Two Lovers"
	Register(Spec{
		OracleID:     "c2df54f1-a2db-4de3-bc96-5c07f4594752",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create two 1/1 white Ally creature tokens"),
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 2})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "search your library for a Mountain or Cave card"),
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.HasSubtypeWord("Mountain") || c.HasSubtypeWord("Cave")
						},
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "The Cave of Two Lovers — a Mountain or Cave card",
					}.Apply(NewContext(g, item))
				}),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "earthbend 3"), EarthbendTargets(),
				func(g *game.Game, item *game.StackItem) error {
					return EarthbendFirstTarget(EarthbendCount(3))(item, NewContext(g, item))
				}),
		},
	})
}
