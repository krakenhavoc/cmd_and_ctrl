package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prosper, Tome-Bound — Legendary Creature — Tiefling Warlock
// {2}{B}{R}, 1/4 (EDHREC rank 2175):
//
//	"Deathtouch
//	 Mystic Arcanum — At the beginning of your end step, exile the top
//	 card of your library. Until the end of your next turn, you may
//	 play that card.
//	 Pact Boon — Whenever you play a card from exile, create a
//	 Treasure token."
//
// The impulse-exile commander. Deathtouch rides PrintedKeywords.
// Mystic Arcanum is an end-step trigger on the controller's own end
// step that exiles the top card with a play grant lasting until the
// end of the controller's NEXT turn — the grant is stamped two rounds
// out as a backstop and a delayed trigger at the next upkeep pins it
// to that turn (b20ExileTopUntilEndOfNextTurn). Pact Boon watches
// two kinds on one ability: a spell cast from exile (EventCast
// carries the origin — the Appa shape) and a land played from exile
// (the EventZoneMove b20LandPlayed reads). Any grant counts, not
// just Prosper's own — a card an opponent's Ragavan handed you, a
// warped creature recast — as printed.
//
// Sandbox simplification, declared (Horn of Greed's): a land played
// from exile is told apart from a land an effect returned from
// exile by the engine's land-drop tally, and when a land already
// came back from exile or a graveyard by an effect earlier in the
// same turn the tally is ambiguous and the play makes no Treasure.
// Weaker than printed, never stronger. A spell cast from exile is
// always seen.
func init() {
	Register(Spec{
		OracleID:        "1e9b0fd6-5aae-401a-8df5-d94ce6442696",
		Name:            "Prosper, Tome-Bound",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"A land played from exile makes no Treasure if another land already came back to the battlefield from exile or a graveyard earlier that turn."},
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Prosper, Tome-Bound — exile the top card of your library; you may play it until the end of your next turn",
						func(g *game.Game, item *game.StackItem) error {
							return b20ExileTopUntilEndOfNextTurn(g, item, 1,
								"Prosper, Tome-Bound — the exiled card may be played until end of turn")
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventCast, game.EventZoneMove},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b20PlayedACardFromExile(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Prosper, Tome-Bound — create a Treasure",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
