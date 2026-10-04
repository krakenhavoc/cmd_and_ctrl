package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Erasure — Sorcery {U}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Surveil 1."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then
// surveil 1 (CR 701.25a) on the next line (ADR 0116 §6). The surveil
// is a prompt of its own, queued behind the pick, so it is answered
// after the discard as printed (CR 608.2c). You surveil even when
// nothing is discarded (the 2018-10-05 ruling; CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bae4e34-fcf4-4aee-8da2-5b3ee41a595a",
		Name:         "Thought Erasure",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
			}).Apply(ctx); err != nil {
				return err
			}
			return Surveil{N: 1}.Apply(ctx)
		},
	})
}
