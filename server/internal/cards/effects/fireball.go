package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fireball — Sorcery {X}{R}:
//
//	"This spell costs {1} more to cast for each target beyond the
//	 first.
//	 Fireball deals X damage divided evenly, rounded down, among any
//	 number of targets."
//
// #746: the surcharge is a self cost modifier that reads the announced
// targets (CostsMorePerTargetBeyondFirst sets ReadsTargets), so the
// cast pays {X}{R} at one target, {X}{1}{R} at two and {X}{2}{R} at
// three, and the legal-move enumerator prices each target set on its
// own (ADR 0048 addendum §14).
//
// The rulings this follows:
//   - zero targets is a legal cast that deals no damage;
//   - the damage is divided as Fireball resolves, among the targets
//     still legal then, not as it is cast;
//   - with more legal targets than X, each share rounds down to zero
//     and nothing is dealt.
//
// The X picker opens before targeting, so it prices at one target and
// quotes the surcharge clause under its readout (ADR 0048 addendum,
// open question 2).
func init() {
	Register(Spec{
		OracleID:     "aa7714b0-2bfb-458a-8ebf-37ec2c53383e",
		Name:         "Fireball",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{1}", "This spell costs {1} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			share := ctx.X() / len(legal)
			if share <= 0 {
				return nil
			}
			for _, t := range legal {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: share}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
