package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Intervention — Sorcery {4}{R} (Dominaria, #5530):
//
//	"Choose one —
//	 • Fiery Intervention deals 5 damage to target creature.
//	 • Destroy target artifact."
//
// Abrade's two modes, with 5 damage on the first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8cd3a52-1056-4664-ae55-eb0ba3a43a15",
		Name:         "Fiery Intervention",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Fiery Intervention deals 5 damage to target creature.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 5))),
			effects.Mode("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if ctx.HasMode(0) {
				return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 5}.Apply(ctx)
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
