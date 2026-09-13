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
// Sandbox simplification, the Garruk's Uprising posture: the
// intervening-if is checked when the trigger would go on the stack
// and not again on resolution. Playing a land in response to make
// the counts equal should make the trigger do nothing; here the
// search still happens. A corner case, and the direction is
// stronger, so it is declared rather than hidden: no re-check hook on
// a built StackItem exists.
func init() {
	Register(Spec{
		OracleID:     "d2d9ecea-7925-420e-98b9-2f87f41f387c",
		Name:         "Land Tax",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The land-count check happens only when the trigger goes on the stack, so playing a land in response won't stop the search."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b03OpponentControlsMoreLands(g, source.Controller)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Land Tax — search for up to three basic lands",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player:    item.Controller,
							Predicate: IsBasicLand,
							Dest:      game.ZoneHand,
							Limit:     3,
							Reveal:    true,
							Shuffle:   true,
							Reason:    "Land Tax — up to three basic land cards, to your hand",
						}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Land Tax — search your library for up to three basic land cards?",
			},
		}},
	})
}
