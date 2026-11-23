package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Declare Dominance — Sorcery {3}{G}{G} (Jumpstart 2022, #5948):
//
//	"Target creature gets +3/+3 until end of turn. All creatures able to
//	 block it this turn do so."
//
// A +3/+3 pump, then Bloodscent's lure on the same target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db640c5b-4548-44b1-a9cd-cc6838707dbb",
		Name:         "Declare Dominance",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Declare Dominance — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.LureTargetForTheTurn(ctx, "Declare Dominance — all creatures able to block it do so")
		},
	})
}
