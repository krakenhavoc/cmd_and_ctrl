package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kyoshi Island Plaza — Legendary Enchantment — Shrine {3}{G}:
//
//	"When Kyoshi Island Plaza enters, search your library for up to X
//	 basic land cards, where X is the number of Shrines you control. Put
//	 those cards onto the battlefield tapped, then shuffle.
//	 Whenever another Shrine you control enters, search your library for
//	 a basic land card, put it onto the battlefield tapped, then shuffle."
//
// X is counted as the enters trigger resolves (the ruling), the Plaza
// itself included while it is still a Shrine you control. With X = 0 the
// library is still searched and shuffled but nothing can be found
// (Traverse the Outlands's shape). The Shrine type is read through the
// layered view, of the entering permanent and of each one counted.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "c0cdc333-6c5a-401c-9810-220d2027f2fe",
		Name:         "Kyoshi Island Plaza",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kyoshi Island Plaza — search for up to X basic lands, tapped (X = Shrines you control)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := kyoshiIslandPlazaShrines(g, item.Controller)
					return kyoshiIslandPlazaSearch(ctx, x, "Kyoshi Island Plaza — up to X basic lands, tapped")
				}),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.HasSubtypeWord("Shrine")
			}, "Kyoshi Island Plaza — search for a basic land, tapped (another Shrine entered)",
				func(g *game.Game, item *game.StackItem) error {
					return kyoshiIslandPlazaSearch(NewContext(g, item), 1, "Kyoshi Island Plaza — a basic land, tapped")
				}),
		},
	})
}

// kyoshiIslandPlazaShrines counts the Shrines a player controls.
func kyoshiIslandPlazaShrines(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtypeWord("Shrine") {
			n++
		}
	}
	return n
}

// kyoshiIslandPlazaSearch is "search your library for up to n basic land
// cards, put them onto the battlefield tapped, then shuffle". A stated
// quality may always find fewer (CR 701.23b); n = 0 still searches and
// shuffles.
func kyoshiIslandPlazaSearch(ctx *Context, n int, reason string) error {
	match := IsBasicLand
	if n <= 0 {
		match = func(game.Card) bool { return false }
	}
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     match,
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        reason,
		Source:        ctx.Source(),
	}.Apply(ctx)
}
