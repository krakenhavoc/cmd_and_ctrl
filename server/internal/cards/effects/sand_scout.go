package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ca9f9017-c952-4e89-b284-02396e6be0af",
		Name:         "Sand Scout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b05OpponentControlsMoreLands(g, source.Controller)
			}, "Sand Scout — search for a Desert card", func(g *game.Game, item *game.StackItem) error {
				if !b05OpponentControlsMoreLands(g, item.Controller) {
					return nil // CR 603.4: re-checked on resolution
				}
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsLandWithSubtype("desert"),
					Dest:          game.ZoneBattlefield,
					TappedOnEntry: true,
					Limit:         1,
					Shuffle:       true,
					Reason:        "Sand Scout — a Desert card",
				}.Apply(NewContext(g, item))
			}),
			TriggersOnlyOnceEachTurn(OncePerBatch(OnAny(cmdr2S14GraveyardEvents, gitrogLandToGraveyard,
				"Sand Scout — create a 1/1 red, green, and white Sand Warrior creature token",
				Do(CreateToken{Template: TokenCard("1/1 red, green, and white Sand Warrior"), N: 1})))),
		},
	})
}
