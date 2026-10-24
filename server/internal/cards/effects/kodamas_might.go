package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kodama's Might — Instant — Arcane {G}:
//
//	"Target creature gets +2/+2 until end of turn.
//	Splice onto Arcane {G}"
func init() {
	Register(Spec{
		OracleID:     "c027a8ef-6cdd-4c14-9d4f-b21c63312874",
		Name:         "Kodama's Might",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{G}"),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Kodama's Might — +2/+2"}.Apply(ctx)
		},
	})
}
