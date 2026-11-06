package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9dbc2fc8-ed42-4534-9676-bf0695cd518c",
		Name:         "You're Not Alone",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2
			if b04CreaturesControlled(ctx.Game, ctx.Controller()) >= 3 {
				n = 4
			}
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: n, Toughness: n, Label: "You're Not Alone — gets bigger"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
