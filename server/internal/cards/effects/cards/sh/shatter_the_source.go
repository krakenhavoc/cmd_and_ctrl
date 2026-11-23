package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shatter the Source — Instant {5}{R} (March of the Machine):
//
//	"Convoke. Choose one —
//	 • Shatter the Source deals 6 damage to target creature, planeswalker,
//	   or battle.
//	 • Destroy target artifact."
//
// Abrade's two modes with a wider first target and 6 damage. The creature,
// planeswalker or battle filter is Or(Creature(), Planeswalker(), Battle()).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8b3bcf2-1be1-46ed-9531-ca4ce98e3b3c",
		Name:         "Shatter the Source",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Shatter the Source deals 6 damage to target creature, planeswalker, or battle.",
				effects.TargetPermanent("target creature, planeswalker, or battle", effects.Or(effects.Creature(), effects.Planeswalker(), effects.Battle()))),
				effects.ForTargets(effects.DamageToTarget(0, 6))),
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
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 6}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
