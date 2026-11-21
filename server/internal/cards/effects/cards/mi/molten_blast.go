package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5a680186-a84d-400f-9452-76f19bdfafa3",
		Name:         "Molten Blast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Molten Blast deals 2 damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
			effects.Mode("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if ctx.HasMode(0) {
				return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 2}.Apply(ctx)
			}
			return effects.DestroyTarget{Target: id}.Apply(ctx)
		},
	})
}
