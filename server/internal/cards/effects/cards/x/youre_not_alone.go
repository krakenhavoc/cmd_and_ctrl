package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You're Not Alone — Instant {W} (Final Fantasy, #4462):
//
//	"Target creature gets +2/+2 until end of turn. If you control three
//	 or more creatures, it gets +4/+4 until end of turn instead."
//
// Olivia's Midnight Ambush's shape: "instead" is one read of the
// creature count as the spell resolves, so the target gets one pump of
// whichever size, never both. The target counts toward the three when
// you control it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9dbc2fc8-ed42-4534-9676-bf0695cd518c",
		Name:         "You're Not Alone",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 2
			if effects.B04CreaturesControlled(ctx.Game, ctx.Controller()) >= 3 {
				n = 4
			}
			for _, t := range ctx.LegalTargets() {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: n, Toughness: n, Label: "You're Not Alone — gets bigger"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
