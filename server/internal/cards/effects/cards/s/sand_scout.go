package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sand Scout:
//
//	"When this creature enters, if an opponent controls more lands than
//	 you, search your library for a Desert card, put it onto the
//	 battlefield tapped, then shuffle.
//	 Whenever one or more land cards are put into your graveyard from
//	 anywhere, create a 1/1 red, green, and white Sand Warrior creature
//	 token. This ability triggers only once each turn."
//
// The enters trigger is an intervening "if" (CR 603.4): checked as it
// triggers and again as it resolves. The graveyard trigger is a "one or
// more" batch (CR 603.2c), limited to once each turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca9f9017-c952-4e89-b284-02396e6be0af",
		Name:         "Sand Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.B05OpponentControlsMoreLands(g, source.Controller)
			}, "Sand Scout — search for a Desert card", func(g *game.Game, item *game.StackItem) error {
				if !effects.B05OpponentControlsMoreLands(g, item.Controller) {
					return nil // CR 603.4: re-checked on resolution
				}
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsLandWithSubtype("desert"),
					Dest:          game.ZoneBattlefield,
					TappedOnEntry: true,
					Limit:         1,
					Shuffle:       true,
					Reason:        "Sand Scout — a Desert card",
				}.Apply(effects.NewContext(g, item))
			}),
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.OnAny(effects.Cmdr2S14GraveyardEvents, effects.GitrogLandToGraveyard,
				"Sand Scout — create a 1/1 red, green, and white Sand Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red and green and white Sand Warrior"), N: 1})))),
		},
	})
}
