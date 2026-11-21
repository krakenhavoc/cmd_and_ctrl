package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chronicler of Heroes — Creature — Centaur Wizard {1}{G}{W}, 3/3
// (Ravnica Allegiance):
//
//	"When this creature enters, draw a card if you control a creature with
//	 a +1/+1 counter on it."
//
// Not an intervening "if" (CR 603.4): the trigger always goes on the stack
// and the board is looked at as it resolves, so a counter put on a creature
// in response counts. The Chronicler itself counts if it has a +1/+1 counter
// on it (an entry-counter replacement), and a creature you control with
// any other counter does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3779271-1186-4754-8b8d-0ee18e585bb6",
		Name:         "Chronicler of Heroes",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Chronicler of Heroes — draw a card if you control a creature with a +1/+1 counter on it",
				effects.Do(effects.If{
					Cond: effects.CondBoard{
						Name: "you control a creature with a +1/+1 counter on it",
						Test: func(ctx *effects.Context) bool {
							for _, c := range ctx.Game.BattlefieldCardsForEffect() {
								if c.Controller == ctx.Controller() && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
									return true
								}
							}
							return false
						},
					},
					Then: []effects.Applier{effects.DrawCards{N: 1}},
				})),
		},
	})
}
