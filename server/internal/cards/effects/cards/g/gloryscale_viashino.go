package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloryscale Viashino — Creature — Lizard Soldier {1}{R}{G}{W}, 3/3
// (Alara Reborn, #5586):
//
//	"Whenever you cast a multicolored spell, this creature gets +3/+3
//	 until end of turn."
//
// Hero of Precinct One's cast trigger (b25CastIsMulticolored reads the
// spell's colours as it is cast), with a self pump that ends in the
// cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d94a6e12-0197-4125-9a8c-431e42edb1da",
		Name:         "Gloryscale Viashino",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && effects.B25CastIsMulticolored(ev, g)
			}, "Gloryscale Viashino — this creature gets +3/+3 until end of turn",
				effects.ThisGetsUntilEndOfTurn(3, 3, "Gloryscale Viashino — +3/+3 until end of turn")),
		},
	})
}
