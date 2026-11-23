package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Start from Scratch — Sorcery — Lesson {2}{R} (Strixhaven: School of
// Mages, #5502):
//
//	"Choose one —
//	 • Start from Scratch deals 1 damage to any target.
//	 • Destroy target artifact."
//
// Abrade's two bullets with "any target" (a creature, a player, a
// planeswalker or a battle) for the damage. Each bullet carries its own
// target clause, judged at announce and again as the spell resolves.
// Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8e255ec-7d66-43f6-b5f0-712f5957687f",
		Name:         "Start from Scratch",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Start from Scratch deals 1 damage to any target.", effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
			effects.Mode("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0]
			switch {
			case ctx.HasMode(0):
				return effects.DealDamage{Source: ctx.Source(), Target: target.ID, Amount: 1}.Apply(ctx)
			case ctx.HasMode(1):
				if target.Kind != game.TargetCard {
					return nil
				}
				return effects.DestroyTarget{Target: target.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
