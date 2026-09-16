package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of the White Orchid — Creature — Human Knight {W}{W}, 2/2
// (EDHREC rank 647):
//
//	"First strike
//	 When this creature enters, if an opponent controls more lands
//	 than you, you may search your library for a Plains card, put it
//	 onto the battlefield, then shuffle."
//
// White's catch-up ramp, and in a four-player game the condition is
// nearly always true for whoever is behind. First strike rides
// PrintedKeywords. The ETB is an intervening-if "you may" trigger:
// the land count is compared in AppliesTo (no prompt at all when you
// are not behind — CR 603.4), the "you may" is the ordinary optional
// prompt, and the count is checked AGAIN at resolution, so a land
// drop in response that catches you up makes the trigger do nothing,
// as printed.
//
// "A Plains card" is the land TYPE — a Snow-Covered Plains or a
// Savannah qualifies — and it enters UNTAPPED (the card does not say
// tapped, unlike Knight of the Reliquary's cousins).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7d0efc64-151f-45b4-8c0b-c32609f9d862",
		Name:            "Knight of the White Orchid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b05OpponentControlsMoreLands(g, source.Controller)
			}, "Knight of the White Orchid — search for a Plains", func(g *game.Game, item *game.StackItem) error {
				if !b05OpponentControlsMoreLands(g, item.Controller) {
					return nil // CR 603.4: re-checked on resolution
				}
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsLandWithSubtype("plains"),
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Knight of the White Orchid — a Plains card",
				}.Apply(NewContext(g, item))
			}), "Knight of the White Orchid — search your library for a Plains card and put it onto the battlefield?"),
		},
	})
}
