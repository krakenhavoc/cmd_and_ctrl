package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a3e51a35-09df-4189-b131-08a21e6a557d",
		Name:         "Cryptic Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Counter target spell.",
				TargetSpell("target spell"),
				CounterTheModesTarget),
			ModeDoing("Return target permanent to its owner's hand.",
				TargetPermanent("target permanent"),
				BounceTheModesTarget),
			ModeDoing("Tap all creatures your opponents control.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return tapAllCreaturesYourOpponentsControl(ctx.Game, item)
				}),
			ModeWithPurpose(ModeDoing("Draw a card.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}), game.Purpose{Draws: 1, Answers: game.AnswerValue}),
		),
	})
}
