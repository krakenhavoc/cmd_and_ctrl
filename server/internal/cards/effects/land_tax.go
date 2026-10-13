package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Land Tax — Enchantment {W} (EDHREC rank 367):
//
//	"At the beginning of your upkeep, if an opponent controls more
//	 lands than you, you may search your library for up to three
//	 basic land cards, reveal them, put them into your hand, then
//	 shuffle."
//
// White's card advantage: three cards an upkeep for one mana, as
// long as someone at the table has more lands — which at a
// four-player table is nearly always. An upkeep trigger with an
// intervening-if (CR 603.4), a "you may" (the trigger prompt), and a
// search to hand with the chooser deciding which basics and how
// many — "up to three" is the chooser's zero minimum.
//
// The intervening-if is checked twice (CR 603.4): when the trigger
// would go on the stack, in AppliesTo, and again when it resolves, at
// the top of the Effect. So a land played in response to even out the
// counts makes the trigger do nothing, as printed. (Until #3138 this
// file declared that second check as missing; it never needed engine
// support, because the Effect already runs at resolution and reads the
// live board.)
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2d9ecea-7925-420e-98b9-2f87f41f387c",
		Name:         "Land Tax",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b03OpponentControlsMoreLands(g, source.Controller)
			}, "Land Tax — search for up to three basic lands", func(g *game.Game, item *game.StackItem) error {
				if !b03OpponentControlsMoreLands(g, item.Controller) {
					return nil // CR 603.4: the condition no longer holds on resolution
				}
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     3,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Land Tax — up to three basic land cards, to your hand",
				}.Apply(NewContext(g, item))
			}), "Land Tax — search your library for up to three basic land cards?"),
		},
	})
}
