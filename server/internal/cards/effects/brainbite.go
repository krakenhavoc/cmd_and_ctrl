package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brainbite — Sorcery {2}{U}{B}:
//
//	"Target opponent reveals their hand. You choose a card from it.
//	 That player discards that card.
//	 Draw a card."
//
// The revealed-hand pick (ADR 0116) with no filter, then the draw on
// the next line, as Thoughtseize's life loss runs (ADR 0116 §6). The
// candidates are fixed at the reveal, so the draw cannot change what
// may be chosen, and it reads nothing about the pick. You draw even
// when the opponent's hand is empty (the 2009-05-01 rulings; CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fef94125-aa8d-4147-a609-1e990961bde2",
		Name:         "Brainbite",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{Player: TargetedPlayer(ctx), Label: "card"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
