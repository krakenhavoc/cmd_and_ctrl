package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temur Battle Rage — Instant {1}{R} (Commander Masters, #5948):
//
//	"Target creature gains double strike until end of turn.
//	 Ferocious — That creature also gains trample until end of turn if you
//	 control a creature with power 4 or greater."
//
// Ferocious is read as the spell resolves (CR 608.2), over the layered
// power of the creatures its controller then controls, so a creature
// pumped or killed in response counts as it stands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fdc0d17-b683-4263-949e-4f74ec76badf",
		Name:         "Temur Battle Rage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			keywords := []string{"double strike"}
			if b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 4) {
				keywords = append(keywords, "trample")
			}
			return GrantKeywordUntilEOT{Target: target, Keywords: keywords, Label: "Temur Battle Rage — double strike"}.Apply(ctx)
		},
	})
}
