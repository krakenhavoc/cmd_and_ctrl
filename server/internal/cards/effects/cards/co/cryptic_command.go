package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryptic Command — Instant {1}{U}{U}{U}:
//
//	"Choose two —
//	 • Counter target spell.
//	 • Return target permanent to its owner's hand.
//	 • Tap all creatures your opponents control.
//	 • Draw a card."
//
// Two different modes (ChooseN, not repeating), both chosen as it is
// cast (CR 700.2). With a targeted mode chosen and every target
// illegal, the whole spell does nothing (CR 608.2b, the 2017 ruling):
// the engine's fizzle check covers the untargeted mode too.
//
// No simplification. Star Trek Commander (#4974).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3e51a35-09df-4189-b131-08a21e6a557d",
		Name:         "Cryptic Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Counter target spell.",
				effects.TargetSpell("target spell"),
				effects.CounterTheModesTarget),
			effects.ModeDoing("Return target permanent to its owner's hand.",
				effects.TargetPermanent("target permanent"),
				effects.BounceTheModesTarget),
			effects.ModeDoing("Tap all creatures your opponents control.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.TapAllCreaturesYourOpponentsControl(ctx.Game, item)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Draw a card.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}), game.Purpose{Draws: 1, Answers: game.AnswerValue}),
		),
	})
}
