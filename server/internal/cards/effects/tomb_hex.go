package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomb Hex — Instant {2}{B} (Worldwake, #5582):
//
//	"Target creature gets -2/-2 until end of turn.
//	 Landfall — If you had a land enter the battlefield under your control
//	 this turn, that creature gets -4/-4 until end of turn instead."
//
// "If you had a land enter the battlefield under your control this
// turn" reads the controller's turn tally (wwkHadLandfallThisTurn),
// which counts every land that entered under them this turn, one that
// has since left included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b3bc59c-5439-4ec8-b25e-7493fa1cd3fd",
		Name:         "Tomb Hex",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -2, -2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 2
			if wwkHadLandfallThisTurn(ctx) {
				n = 4
			}
			return BoostUntilEOT{Target: id, Power: -n, Toughness: -n, Label: "Tomb Hex"}.Apply(ctx)
		},
	})
}
