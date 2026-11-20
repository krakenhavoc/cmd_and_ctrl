package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "861fa80d-99e0-4332-a2b8-5aa959fd41a4",
		Name:         "Consume Spirit",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		SpendOnly:    effects.SpellSpendOnlyOnX("B"),
		Targets:      effects.TargetAny(),
		OnResolve: effects.ResolveSteps(
			// The life gain waits for a prompt the damage asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error { return effects.CmdrB19DealXToTheTarget(ctx) }),
			// "and you gain X life" is not conditional on the target.
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: ctx.X()}.Apply(ctx)
			}),
		),
	})
}
