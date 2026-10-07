package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consume Spirit — Sorcery {X}{1}{B}:
//
//	"Spend only black mana on X.
//	 Consume Spirit deals X damage to any target and you gain X life."
//
// #2556: Drain Life's simpler sibling. The gain is X, not "the damage
// dealt", so a prevented hit still gains the life, exactly as printed.
// SpellSpendOnlyOnX("B") is the cast-side clause (ADR 0040's
// 2026-10-07 amendment).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "861fa80d-99e0-4332-a2b8-5aa959fd41a4",
		Name:         "Consume Spirit",
		XMatters:     true,
		Completeness: CompletenessFull,
		SpendOnly:    SpellSpendOnlyOnX("B"),
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if t, ok := firstLegalTarget(ctx); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}).Apply(ctx); err != nil {
					return err
				}
			}
			// "and you gain X life" is not conditional on the target.
			return GainLife{Player: ctx.Controller(), Amount: ctx.X()}.Apply(ctx)
		},
	})
}
