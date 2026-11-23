package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shredded Sails — Instant {1}{R} (Jumpstart 2022, #5948):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Shredded Sails deals 4 damage to target creature with flying.
//	 Cycling {2}"
//
// Abrade's two-mode shape: each mode has its own target clause, so only the
// chosen mode's target is picked (and the creature must have flying, read
// through the layers), plus the cycling keyword's hand ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea908d2d-d443-4385-a71a-ebb00938cd57",
		Name:         "Shredded Sails",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.ModeWithPurpose(effects.Mode("Shredded Sails deals 4 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))), effects.ForTargets(effects.DamageToTarget(0, 4))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 4}.Apply(ctx)
			}
			return nil
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
