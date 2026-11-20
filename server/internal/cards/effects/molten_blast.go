package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Blast — Instant {2}{R} (Zendikar Rising, #5508):
//
//	"Choose one —
//	 • Molten Blast deals 2 damage to target creature or planeswalker.
//	 • Destroy target artifact."
//
// Abrade's shape: each bullet carries its own target clause, and the
// chosen bullet's target is re-checked as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a680186-a84d-400f-9452-76f19bdfafa3",
		Name:         "Molten Blast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Molten Blast deals 2 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
				ForTargets(DamageToTarget(0, 2))),
			Mode("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if ctx.HasMode(0) {
				return DealDamage{Source: ctx.Source(), Target: id, Amount: 2}.Apply(ctx)
			}
			return DestroyTarget{Target: id}.Apply(ctx)
		},
	})
}
