package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ready to Rumble — Sorcery {4}{R} (Streets of New Capenna, #5491):
//
//	"Choose one —
//	 • Ready to Rumble deals 5 damage to target creature or planeswalker.
//	 • Destroy target artifact."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04dd9395-847b-46f7-a669-9b0df89d01ff",
		Name:         "Ready to Rumble",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Ready to Rumble deals 5 damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))), effects.ForTargets(effects.DamageToTarget(0, 5))),
			effects.Mode("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 5}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
