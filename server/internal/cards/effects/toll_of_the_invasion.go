package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toll of the Invasion — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card.
//	 Amass Zombies 1."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then
// amass Zombies 1 (CR 701.47a) on the next line (ADR 0116 §6): the
// Army is yours and cannot change what may be chosen. You amass even
// when there is no nonland card to take (the 2019-05-03 ruling;
// CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a7c2e32-8585-4f7f-8635-199e3c6cd8a9",
		Name:         "Toll of the Invasion",
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
			return Amass{Subtype: "Zombie", N: 1}.Apply(ctx)
		},
	})
}
