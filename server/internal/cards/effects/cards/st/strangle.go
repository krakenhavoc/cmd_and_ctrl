package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strangle — Sorcery {R} (Streets of New Capenna, #5491):
//
//	"Strangle deals 3 damage to target creature or planeswalker."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f570526-7fb0-4fe5-8014-b5793eb65f87",
		Name:         "Strangle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 3}.Apply(ctx)
		},
	})
}
