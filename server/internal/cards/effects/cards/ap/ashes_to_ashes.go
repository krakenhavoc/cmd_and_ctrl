package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashes to Ashes — Sorcery for {1}{B}{B}:
//
//	"Exile two target nonartifact creatures. Ashes to Ashes deals 5
//	damage to you."
//
// S20 sub-PR 5: exactly two distinct targets sharing one predicate.
// The 5 damage is not contingent on the exiles — it lands whenever
// the spell resolves, including when one target has become illegal
// (only the all-illegal case fizzles the whole spell).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "944a52d9-bb14-43c5-8d05-2afaf023dc9f",
		Name:         "Ashes to Ashes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target nonartifact creatures", effects.Not(effects.Artifact())).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 5}.Apply(ctx)
		},
	})
}
