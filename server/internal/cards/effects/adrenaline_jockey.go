package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adrenaline Jockey — Creature — Minotaur Pilot {2}{R}, 3/3 (Aetherdrift,
// #4464):
//
//	"Whenever a player casts a spell, if it's not their turn, this
//	 creature deals 4 damage to them.
//	 Whenever you activate an exhaust ability, put a +1/+1 counter on
//	 this creature."
//
// The first ability is Scytheclaw Raptor's, word for word: any player's
// cast, its controller's own included, and the intervening "if" is
// whether the caster is the active player (b29SpellCastOffTurn). The
// second is Rangers' Refueler's watch on an exhaust activation
// (WheneverYouActivateAnExhaustAbility), mana abilities included; the
// counter goes on the Jockey only while it is still the same object.
//
// No simplification.
func init() {
	const burn = "Adrenaline Jockey — 4 damage to the player who cast a spell on another player's turn"
	Register(Spec{
		OracleID:     "c3dc6b0e-8790-4bd3-b166-63b2dc77488f",
		Name:         "Adrenaline Jockey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b29SpellCastOffTurn(ev, g)
				},
				Key: burn,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, burn)
					item.Params.Player = ev.Actor
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DealDamage{Source: item.SourceCardID, Target: item.Params.Player, Amount: 4}.Apply(NewContext(g, item))
				},
			},
			WheneverYouActivateAnExhaustAbility("Adrenaline Jockey — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
