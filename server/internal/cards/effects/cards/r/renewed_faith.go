package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renewed Faith — Instant {2}{W} (Dominaria Remastered, #5948):
//
//	"You gain 6 life.
//	 Cycling {1}{W}
//	 When you cycle this card, you may gain 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa4ecefa-808e-4706-af16-4ea5af9743fc",
		Name:         "Renewed Faith",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 6},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: 6}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Renewed Faith — cycled: you may gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(effects.NewContext(g, item))
					})),
				"Renewed Faith — gain 2 life?"),
		},
	})
}
