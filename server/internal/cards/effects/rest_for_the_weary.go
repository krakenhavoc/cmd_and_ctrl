package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rest for the Weary — Instant {1}{W} (Worldwake, #5582):
//
//	"Target player gains 4 life.
//	 Landfall — If you had a land enter the battlefield under your control
//	 this turn, that player gains 8 life instead."
//
// "If you had a land enter the battlefield under your control this
// turn" reads the controller's turn tally (wwkHadLandfallThisTurn),
// which counts every land that entered under them this turn, one that
// has since left included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7f08962-6c54-4fd5-a55e-7513e758af32",
		Name:         "Rest for the Weary",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 4}),
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			p, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			n := 4
			if wwkHadLandfallThisTurn(ctx) {
				n = 8
			}
			return GainLife{Player: p, Amount: n}.Apply(ctx)
		},
	})
}
