package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Endstone — Legendary Artifact {7} (EDHREC rank 3315):
//
//	"Whenever you play a land or cast a spell, draw a card.
//	 At the beginning of your end step, your life total becomes half
//	 your starting life total, rounded up."
//
// Seven mana for a card off every land and every spell, paid for in
// life every end step. Two triggers:
//
//   - The draw is ONE printed ability with two trigger conditions,
//     watching the land play and the cast on one declaration
//     (b31YouPlayedALandOrCastASpell). A cast is EventCast by the
//     controller; a land play is b20LandPlayed's read of the zone
//     move — a hand origin is always a play, and a play from exile
//     or a graveyard is told from a returned land by the land-drop
//     tally, with the declared ambiguity that helper documents. The
//     Endstone itself is on the stack when it is cast, so it never
//     draws for its own casting.
//   - The end step sets the controller's life to half the format's
//     starting total rounded up — 20 in Commander — the CR 119.5
//     way: the player gains or loses the difference, so a lifegain
//     or life-loss payoff sees the change, and a total already at
//     20 changes nothing (b31LifeBecomes).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57ee0730-3ac5-4f17-ac82-bd4a6673c2bb",
		Name:         "The Endstone",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventZoneMove, game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b31YouPlayedALandOrCastASpell(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "The Endstone — draw a card",
						func(g *game.Game, item *game.StackItem) error {
							return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "The Endstone — your life total becomes half your starting life total",
						func(g *game.Game, item *game.StackItem) error {
							return b31LifeBecomes(NewContext(g, item), item.Controller, (game.StartingLife+1)/2)
						})
				},
			},
		},
	})
}
