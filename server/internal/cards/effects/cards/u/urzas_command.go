package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Command — Instant {2}{U}{U} (The Brothers' War Commander, #5482):
//
//	"Choose two —
//	 • Creatures you don't control get -2/-0 until end of turn.
//	 • Create a tapped Powerstone token.
//	 • Create a tapped 0/0 colorless Construct artifact creature token
//	   with 'This token gets +1/+1 for each artifact you control.'
//	 • Scry 1, then draw a card."
//
// Kayla's Command's "choose two" template. The bullets run in printed
// order (CR 608.2c): a tapped Powerstone made by the second bullet is an
// artifact the Construct of the third counts. The -2/-0 set is the
// creatures you don't control as the spell resolves (CR 611.2c). The
// Powerstone is the printed token and the Construct is Urza's
// (UrzaConstructToken), both entering tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5c30dec-072c-4f02-b5d8-5eccfd1cffcf",
		Name:         "Urza's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Creatures you don't control get -2/-0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.OpponentControls()),
						Power: -2,
						Label: "Urza's Command — creatures you don't control get -2/-0",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerValue}),
			effects.ModeWithPurpose(effects.ModeDoing("Create a tapped Powerstone token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1}),
			effects.ModeWithPurpose(effects.ModeDoing("Create a tapped 0/0 colorless Construct artifact creature token with \"This token gets +1/+1 for each artifact you control.\"", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateTokenAdvanced{Spec: effects.Token(effects.UrzaConstructToken()).EntersTapped(), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1, Answers: game.AnswerMakesBlocker}),
			effects.ModeWithPurpose(effects.ModeSteps("Scry 1, then draw a card.", nil,
				effects.Scry{N: 1}, effects.DrawCards{N: 1}),
				game.Purpose{Scry: 1, Draws: 1, Answers: game.AnswerValue}),
		),
	})
}
