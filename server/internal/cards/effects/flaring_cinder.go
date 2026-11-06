package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flaring Cinder — Creature — Elemental Sorcerer {1}{U/R}{U/R}, 3/2
// (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters and whenever you cast a spell with mana
//	 value 4 or greater, you may discard a card. If you do, draw a card."
//
// One printed ability with two trigger conditions, so one
// TriggeredAbility watching two event kinds (Up the Beanstalk's shape):
// its own entry, or its controller casting a spell whose mana value on the
// stack is 4 or more (X counts at its announced value). The rummage is
// Rubble Rouser's: the discard first, and a draw only for a card
// discarded.
//
// No simplification.
func init() {
	const label = "Flaring Cinder — you may discard a card; if you do, draw a card"
	bigSpell := YouCast(ManaValueGE(4))
	Register(Spec{
		OracleID:     "f70d65bf-5f1c-4aee-9bd1-f5ee6974b058",
		Name:         "Flaring Cinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventCast}, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				if ev.Kind == game.EventETB {
					return ev.CardID == source.InstanceID
				}
				return bigSpell(ev, source, lki, g)
			}, label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
