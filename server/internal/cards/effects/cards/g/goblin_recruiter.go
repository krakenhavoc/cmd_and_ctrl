package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Recruiter — Creature — Goblin {1}{R}, 1/1 (Mystery Booster 2):
//
//	"When this creature enters, search your library for any number of Goblin
//	 cards, reveal them, then shuffle and put those cards on top in any order."
//
// Congregation at Dawn's search with no limit: every Goblin card in the
// library may be taken (the limit is the library's size), revealed, the
// library shuffled and the found cards put on top, then the controller
// arranges them. A search with a stated quality may find fewer, or none
// (CR 701.23b); with fewer than two found there is nothing to order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "792e1072-00bf-4f63-b956-5f20cdf09f11",
		Name:         "Goblin Recruiter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Goblin Recruiter — search for any number of Goblin cards, reveal them, shuffle and put them on top in any order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					limit := 0
					if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Library != nil {
						limit = p.Library.Size()
					}
					if limit == 0 {
						return nil
					}
					return effects.SearchLibrary{
						Player: item.Controller,
						// "Goblin" is a word of the source's text; SearchLibrary lends it to each card (CR 612.2, #3984).
						Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Goblin") },
						Dest:      game.ZoneLibrary,
						ToTop:     true,
						Limit:     limit,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Goblin Recruiter — any number of Goblin cards",
						ThenTail:  ctx.ItemTail(effects.S82GapB23RecruiterOrder),
					}.Apply(ctx)
				}),
		},
	})
}
