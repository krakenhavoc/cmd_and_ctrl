package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akki Ronin — Creature — Goblin Samurai {1}{R}, 1/3 (Kamigawa: Neon
// Dynasty):
//
//	"Whenever a Samurai or Warrior you control attacks alone, you may
//	 discard a card. If you do, draw a card."
//
// Battlefield Scavenger's rummage: the discard first, a draw only for a
// card discarded.
//
// No simplification.
func init() {
	const label = "Akki Ronin — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "4b4c76e1-c480-4d10-93d4-483ef2debc4b",
		Name:         "Akki Ronin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.NeoB09SamuraiOrWarriorAttacksAlone, label,
				func(g *game.Game, item *game.StackItem) error {
					return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
				}),
		},
	})
}
