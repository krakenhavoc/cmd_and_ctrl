package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Might of Old Krosa — Instant {G} (Time Spiral, #5596):
//
//	"Target creature gets +2/+2 until end of turn. If you cast this spell
//	 during your main phase, that creature gets +4/+4 until end of turn
//	 instead."
//
// One pump of +2/+2, or +4/+4 when it was cast in its controller's main
// phase (tspB2MainPhaseAmount, read as it resolves).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b50a6db0-16ce-4e98-afbd-ec37e1653e7c",
		Name:         "Might of Old Krosa",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := tspB2MainPhaseAmount(ctx, 2, 4)
			return BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Might of Old Krosa — pump until end of turn"}.Apply(ctx)
		},
	})
}
