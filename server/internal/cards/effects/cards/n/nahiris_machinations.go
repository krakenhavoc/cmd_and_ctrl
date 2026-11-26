package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nahiri's Machinations — Enchantment {1}{W} (Shadows over Innistrad, S82
// OLD):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gains indestructible until end of turn.
//	 {1}{R}: This enchantment deals 1 damage to target blocking creature."
//
// Aethershield Artificer's combat trigger with the keyword alone, and Balduvian
// Trading Post's "1 damage to target attacking creature" over a blocker.
//
// No simplification.
func init() {
	const label = "Nahiri's Machinations — target creature you control gains indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "b119e0de-6af4-4b5b-9920-187f0a414231",
		Name:         "Nahiri's Machinations",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(label, effects.EbTargetCreatureGainsUntilEOT("indestructible", label)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This enchantment deals 1 damage to target blocking creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.Harmful(effects.TargetCreature("target blocking creature", effects.BlockingCreature())),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
