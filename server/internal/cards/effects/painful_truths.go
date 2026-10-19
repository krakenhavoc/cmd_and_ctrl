package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Painful Truths — Sorcery {2}{B} (EDHREC rank 1006):
//
//	"Converge — You draw X cards and lose X life, where X is the
//	 number of colors of mana spent to cast this spell."
//
// The first converge card in the catalog, and the proof of #761's
// record: X is the number of COLOURS of mana that paid, read back at
// resolution from the stack item.
//
// Two things make it work, and both are the ADR's rather than the
// card's:
//
//   - Spec.WantsDistinctColors tells the cast gate to pay the generic
//     half of {2}{B} with colours it has not spent yet, instead of
//     the usual colourless-first order. Cast off a Swamp, an Island
//     and a Forest, this draws three; cast off a Swamp and two Sol
//     Ring mana, it draws one, which is also correct.
//   - For a payment the engine waived (an unfunded permissive cast or
//     Cast anyway), the caster is asked what they paid outside the
//     pool (#3855), so ctx.ColorsSpentCount() counts every colour
//     they paid.
//
// CR 702.86 makes converge an ability word, so the counting lives on
// the card; the drawing and the life loss are one instruction, so a
// player who converges for five at four life draws five and then dies
// to the CR 704.5a state-based check, as printed.
//
// It declares ReadsManaSpent, which is what makes a waived cast ask.
func init() {
	Register(Spec{
		OracleID:            "58a2d15f-1b4b-4303-b524-8cc0d1b3e7d2",
		Name:                "Painful Truths",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        CompletenessFull,
		WantsDistinctColors: true,
		OnResolve: ResolveSteps(
			// The life loss waits for a prompt the draws ask (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				return DrawCards{N: ctx.ColorsSpentCount()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Amount: -ctx.ColorsSpentCount()}.Apply(ctx)
			}),
		),
	})
}
