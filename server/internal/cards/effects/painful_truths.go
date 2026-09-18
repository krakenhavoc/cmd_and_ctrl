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
//   - ctx.ColorsSpentCount() answers 0 for a payment the engine
//     waived (permissive mode, a strict-mode override). That is the
//     weaker-than-printed direction: the spell resolves and does
//     nothing rather than guessing five.
//
// CR 702.86 makes converge an ability word, so the counting lives on
// the card; the drawing and the life loss are one instruction, so a
// player who converges for five at four life draws five and then dies
// to the CR 704.5a state-based check, as printed.
//
// One caveat for the player, because it is visible at a normal table:
// with strict mana off, the engine never sees the payment and the
// spell draws nothing.
func init() {
	Register(Spec{
		OracleID:            "58a2d15f-1b4b-4303-b524-8cc0d1b3e7d2",
		Name:                "Painful Truths",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so Painful Truths draws nothing — turn strict mana on for it to count colours."},
		WantsDistinctColors: true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.ColorsSpentCount()
			if x <= 0 {
				return nil
			}
			if err := (DrawCards{Player: ctx.Controller(), N: x}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: -x}.Apply(ctx)
		},
	})
}
