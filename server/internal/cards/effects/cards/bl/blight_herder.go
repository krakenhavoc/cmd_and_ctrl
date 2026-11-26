package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blight Herder — Creature — Eldrazi Processor {5}, 4/5 (Battle for
// Zendikar, #5547):
//
//	"When you cast this spell, you may put two cards your opponents own from
//	 exile into their owners' graveyards. If you do, create three 1/1
//	 colorless Eldrazi Scion creature tokens. They have "Sacrifice this
//	 token: Add {C}.""
//
// A cast trigger (CR 603.2). The pick is exactly two cards or none (a set
// rule on the prompt); with fewer than two such cards in exile nothing can be
// done. Tokens are never candidates (CR 111.8). The Scions are the catalog's
// own, printed with their sacrifice ability.
//
// No simplification.
func init() {
	const label = "Blight Herder — you may put two cards your opponents own from exile into their owners' graveyards; if you do, create three 1/1 colorless Eldrazi Scion creature tokens"
	effects.Register(effects.Spec{
		OracleID:     "3159d289-b822-4657-919a-d712623e622e",
		Name:         "Blight Herder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: effects.Self,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Controller, item.Owner = ev.Actor, ev.Actor
				return item
			},
			Effect: effects.S82OldB314BlightHerderLook,
		}},
	})
}
