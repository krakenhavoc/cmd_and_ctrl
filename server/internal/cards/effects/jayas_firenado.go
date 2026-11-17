package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jaya's Firenado — Sorcery {4}{R} (Dominaria United, #5487):
//
//	"Jaya's Firenado deals 5 damage to target creature or planeswalker.
//	 Scry 1."
//
// The damage, then the scry (CR 701.22a). A target that has become illegal
// counters the whole spell, scry included (CR 608.2b).
//
// No simplification.
func init() {
	purpose := ForTargets(DamageToTarget(0, 5))
	purpose.Scry = 1
	Register(Spec{
		OracleID:     "250f804f-b5d5-414c-bb25-476d1fe1f85a",
		Name:         "Jaya's Firenado",
		Completeness: CompletenessFull,
		Purpose:      purpose,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
