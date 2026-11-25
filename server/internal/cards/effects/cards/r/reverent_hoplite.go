package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reverent Hoplite — Creature — Human Soldier {4}{W}, 1/2 (Theros Beyond Death, #5514):
//
//	"When this creature enters, create a number of 1/1 white Human Soldier creature
//	 tokens equal to your devotion to white."
//
// Devotion is counted as the trigger resolves (CR 700.5), the Hoplite's own {W}
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cb6dbb5-9ea2-4c19-8173-148080fb898e",
		Name:         "Reverent Hoplite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Reverent Hoplite — create a number of 1/1 white Human Soldier creature tokens equal to your devotion to white",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.DevotionTo(g, item.Controller, ctx.TextColor("W"))
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Human Soldier"), N: n}.Apply(ctx)
				}),
		},
	})
}
