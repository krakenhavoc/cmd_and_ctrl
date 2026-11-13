package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hearts on Fire — Instant {1}{R} (Foundations Jumpstart, #5464):
//
//	"One or two target creatures each get +2/+1 until end of turn."
//
// Heroic Teamwork without the teamwork: one clause with WithCount(1, 2)
// (CR 601.2c), and the resolution walks ctx.LegalTargets() so a target
// that left in response is skipped rather than erroring (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7502366-1a13-473a-b952-a26ce2e21cf2",
		Name:         "Hearts on Fire",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 1)),
		Targets:      TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: "Hearts on Fire"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
