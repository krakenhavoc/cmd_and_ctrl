package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nahiri's Wrath — Sorcery {2}{R} (Eldritch Moon, S82 OLD, #5541):
//
//	"As an additional cost to cast this spell, discard X cards.
//	 Nahiri's Wrath deals damage equal to the total mana value of the
//	 discarded cards to each of up to X target creatures and/or
//	 planeswalkers."
//
// Channeled Force's discard-X cost (CR 107.3a, 601.2f) with March of
// Swirling Mist's "up to X targets" (CountFromX + UpToX). The damage is the
// discarded cards' mana values summed, read from where the discard put them
// (CR 400.7j); it is dealt to every target still legal as one damage
// instance (CR 608.2b).
//
// No simplification.
func init() {
	targets := effects.TargetPermanent("up to X target creatures and/or planeswalkers",
		effects.Or(effects.Creature(), effects.Planeswalker()))
	targets.CountFromX = true
	targets.UpToX = true
	effects.Register(effects.Spec{
		OracleID:       "420091b5-b966-4233-877d-e9143481e97e",
		Name:           "Nahiri's Wrath",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardXCost(),
		Targets:        targets,
		Purpose:        effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			total := 0
			for _, id := range ctx.Discarded() {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					if mv, ok := ctx.Game.ManaValueForEffect(c); ok {
						total += mv
					}
				}
			}
			return effects.SourceDealsDamageToEachLegalTarget(total)(ctx.Game, item)
		},
	})
}
