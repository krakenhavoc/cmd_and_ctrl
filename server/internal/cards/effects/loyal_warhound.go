package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Warhound — "Vigilance. When this creature enters, if an
// opponent controls more lands than you, search your library for a
// basic Plains card, put it onto the battlefield tapped, then
// shuffle."
//
// The intervening-if (CR 603.4) is checked twice in paper: once
// when the trigger would go on the stack and again on resolution.
// AppliesTo covers the first check; the Effect re-checks so a land
// drop in response correctly fizzles the fetch.
func init() {
	Register(Spec{
		OracleID:        "cc6a83c7-e645-4a53-9550-be79b42cd851",
		Name:            "Loyal Warhound",
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID != source.InstanceID {
					return false
				}
				return anOpponentHasMoreLands(g, source.Controller)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Loyal Warhound — search for a basic Plains",
					func(g *game.Game, item *game.StackItem) error {
						if !anOpponentHasMoreLands(g, item.Controller) {
							return nil
						}
						return SearchLibrary{
							Player:        item.Controller,
							Predicate:     IsBasicLandWithSubtype("plains"),
							Dest:          game.ZoneBattlefield,
							Limit:         1,
							Reveal:        true,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Loyal Warhound — a basic Plains",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
