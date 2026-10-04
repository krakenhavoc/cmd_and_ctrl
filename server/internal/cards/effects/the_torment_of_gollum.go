package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Torment of Gollum — Sorcery {3}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Amass Orcs 2."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then
// amass Orcs 2 (CR 701.47a) on the next line (ADR 0116 §6): the Army is
// yours and cannot change what may be chosen. You amass even when
// nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b70aad1-bcb1-4359-b27e-ba8b40ef2752",
		Name:         "The Torment of Gollum",
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
			return Amass{Subtype: "Orc", N: 2}.Apply(ctx)
		},
	})
}
