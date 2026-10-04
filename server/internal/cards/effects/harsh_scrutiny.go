package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harsh Scrutiny — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a creature card
//	 from it. That player discards that card. Scry 1."
//
// The revealed-hand pick (ADR 0116) filtered to creature cards, then
// scry 1 (CR 701.22a) on the next line (ADR 0116 §6). The scry is a
// prompt of its own, queued behind the pick, so it is answered after
// the discard as printed (CR 608.2c). You scry even when there is no
// creature card to take (the 2016-09-20 ruling; CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31338e24-a437-4587-acab-cec46be021e7",
		Name:         "Harsh Scrutiny",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Creature(),
				Label:  "creature card",
			}).Apply(ctx); err != nil {
				return err
			}
			return Scry{N: 1}.Apply(ctx)
		},
	})
}
