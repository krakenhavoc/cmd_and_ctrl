package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solar Blast — Instant {3}{R} (Dominaria Remastered, #5948):
//
//	"Solar Blast deals 3 damage to any target.
//	 Cycling {1}{R}{R}
//	 When you cycle this card, you may have it deal 1 damage to any target."
//
// The cycling trigger is optional and targeted; the damage is dealt by the
// card itself, from the graveyard (last-known information, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ece3e379-c586-48bd-8742-472b5deccc35",
		Name:         "Solar Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageToFirstTargetStep(3),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.Optional(
					effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Solar Blast — cycled: you may have it deal 1 damage to any target",
						func(g *game.Game, item *game.StackItem) error {
							return effects.DamageToFirstTargetStep(1)(item, effects.NewContext(g, item))
						})),
					"Solar Blast — have it deal 1 damage to any target?"),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
