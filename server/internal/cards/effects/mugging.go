package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mugging — Sorcery {R} (Ravnica Remastered, #5948):
//
//	"Mugging deals 2 damage to target creature. That creature can't block
//	 this turn."
//
// Damage first, then the restriction on the same creature (CR 608.2c). A
// target that has left is skipped (CR 608.2b); one that survives the damage
// cannot block until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e4f08c2-1bad-4a85-8aad-9bb34a5579b1",
		Name:         "Mugging",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetCards(item, ctx.Game)
			if len(ids) == 0 {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: ids[0], Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return RestrictUntilEOT{Target: ids[0], Restrictions: game.CantBlock, Label: "Mugging — can't block"}.Apply(ctx)
		},
	})
}
