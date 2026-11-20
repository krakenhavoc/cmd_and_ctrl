package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glademuse — Creature — Beast {2}{G}, 2/4 (Commander 2020, #5513):
//
//	"Whenever a player casts a spell, if it's not their turn, that player
//	 draws a card."
//
// Scytheclaw Raptor's trigger (b29SpellCastOffTurn: any player, Glademuse's
// controller included, casting on a turn that is not theirs) with a draw for
// the caster in place of the damage. The caster rides Params.Player. Nothing
// between the cast and the trigger's resolution can make it that player's
// turn, so the intervening "if" (CR 603.4) holds on resolution whenever it
// held as the spell was cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "028f2c1d-9c1d-4ac3-adf4-644b5166502b",
		Name:         "Glademuse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29SpellCastOffTurn(ev, g)
			},
			Key: "Glademuse — the player who cast a spell on another player's turn draws a card",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Glademuse — the player who cast a spell on another player's turn draws a card")
				item.Params.Player = ev.Actor
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Params.Player, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
