package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyretic Charge — Sorcery {4}{R} (Outlaws of Thunder Junction Commander):
//
//	"Discard your hand, then draw four cards. For each card discarded this
//	 way, creatures you control get +1/+0 until end of turn.
//	 Plot {3}{R} (...)"
//
// The discard is the whole hand (discardWholeHand) and its count is
// recorded as the first step; the draw and the pump read it. The pump
// locks the creatures you control as the spell finishes, so one that
// arrives later gets nothing (CR 611.2c), and it counts the cards that
// really left the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "01766d9b-b2af-4a27-9e65-08c82e7d7637",
		Name:           "Pyretic Charge",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 4},
		SpecialActions: []game.SpecialAction{Plot("{3}{R}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				n, err := discardWholeHand(ctx.Game, ctx.Controller())
				ctx.RecordStep(pyreticChargeDiscarded, nil, n)
				return err
			}),
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 4}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				n := ctx.StepRecord(pyreticChargeDiscarded).N()
				if n <= 0 {
					return nil
				}
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: n,
					Label: "Pyretic Charge — +1/+0 for each card discarded"}.Apply(ctx)
			}),
		),
	})
}

const pyreticChargeDiscarded = "pyretic-charge/discarded"
