package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
func init() {
	Register(Spec{
		OracleID:     "57ea9f63-8267-4303-a3dd-9ded220294dc",
		Name:         "Rise of the Deathbringer",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Draw cards equal to the greatest power among creatures you control. You lose life equal to the number of cards drawn this way.",
				nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					n := rfSpellBGreatestPowerYouControl(ctx, ctx.Controller())
					if n <= 0 {
						return nil
					}
					if err := (DrawCards{Player: ctx.Controller(), N: n}).Apply(ctx); err != nil {
						return err
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -n)
				}),
			ModeWithPurpose(ModeDoing("All creatures get -3/-3 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: Creature(), Power: -3, Toughness: -3, Label: "Rise of the Deathbringer — -3/-3"}.Apply(ctx)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3}}),
		),
	})
}
