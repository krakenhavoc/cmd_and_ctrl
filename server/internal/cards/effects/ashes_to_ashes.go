package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID: "944a52d9-bb14-43c5-8d05-2afaf023dc9f",
		Name:     "Ashes to Ashes",
		Targets:  TargetCreature("two target nonartifact creatures", Not(Artifact())).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 5}.Apply(ctx)
		},
	})
}
