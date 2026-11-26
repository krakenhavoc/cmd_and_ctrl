package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyline Scout — Creature — Human Scout {1}{W}, 2/1 (Guilds of Ravnica, #5525):
//
//	"Whenever this creature attacks, you may pay {1}{W}. If you do, it gains
//	 flying until end of turn."
//
// Order of the Golden Cricket's MayPay shape at {1}{W}; declining, or a pool
// that cannot pay, grants nothing, and a Scout that left and came back is a
// new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77dfca02-3ca3-4a9c-b3e1-479a951cbfaa",
		Name:         "Skyline Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Skyline Scout — you may pay {1}{W}; if you do, it gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{W}",
						Question: "Skyline Scout — pay {1}{W} to give it flying until end of turn?",
						Then:     skylineScoutPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// skylineScoutPaid grants the keyword. An on-disk key: never renamed, never
// reused.
var skylineScoutPaid = effects.PayUnlessBranches("pay-unless/skyline-scout-flying",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"},
			Label: "Skyline Scout — flying"}.Apply(ctx)
	}, nil)
