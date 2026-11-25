package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evangel of Heliod — Creature — Human Cleric {4}{W}{W}, 1/3 (Kaldheim
// Commander, #5506):
//
//	"When this creature enters, create a number of 1/1 white Soldier
//	 creature tokens equal to your devotion to white."
//
// Devotion (CR 700.5) is counted as the trigger resolves, with Evangel
// itself on the battlefield, so its own {W}{W} counts. Nothing is made at
// zero devotion.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fbe1c02-f5bf-47c7-80de-09a32d01e5db",
		Name:         "Evangel of Heliod",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Evangel of Heliod — create a 1/1 white Soldier for each devotion to white",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.DevotionTo(g, item.Controller, ctx.TextColor("W"))
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: n}.Apply(ctx)
				}),
		},
	})
}
