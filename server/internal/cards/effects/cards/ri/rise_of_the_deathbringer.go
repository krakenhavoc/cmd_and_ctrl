package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise of the Deathbringer — Instant {4}{B} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Draw cards equal to the greatest power among creatures you
//	   control. You lose life equal to the number of cards drawn this way.
//	 • All creatures get -3/-3 until end of turn."
//
// The first bullet reads the greatest power as it resolves and loses that
// much life, the number of cards it draws. The second affects the
// creatures on the battlefield as it resolves (CR 611.2c) and is a
// declared creature sweep.
//
// No simplification.
// riseOfTheDeathbringerRecord files the cards drawn, and so the life lost.
const riseOfTheDeathbringerRecord = "rise-of-the-deathbringer/drawn"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "57ea9f63-8267-4303-a3dd-9ded220294dc",
		Name:         "Rise of the Deathbringer",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			// The life loss waits for a prompt the draw asks (a dredge offer; CR 608.2c, #4155).
			// The number is read once, before the draw, and filed for the loss.
			effects.ModeSteps("Draw cards equal to the greatest power among creatures you control. You lose life equal to the number of cards drawn this way.",
				nil,
				effects.Instruction(func(ctx *effects.Context) error {
					n := effects.RfSpellBGreatestPowerYouControl(ctx, ctx.Controller())
					if n <= 0 {
						return nil
					}
					ctx.RecordStep(riseOfTheDeathbringerRecord, nil, n)
					return effects.DrawCards{N: n}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					n := ctx.StepRecord(riseOfTheDeathbringerRecord).N()
					if n <= 0 {
						return nil
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -n)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("All creatures get -3/-3 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.Creature(), Power: -3, Toughness: -3, Label: "Rise of the Deathbringer — -3/-3"}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3}}),
		),
	})
}
