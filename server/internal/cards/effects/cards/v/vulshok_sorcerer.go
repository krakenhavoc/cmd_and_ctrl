package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulshok Sorcerer — Creature — Human Shaman Sorcerer {1}{R}{R}, 1/1 (Fifth Dawn, #5604):
//
//	"Haste
//	 {T}: This creature deals 1 damage to any target."
//
// Zuran Spellcaster's ping on a hasty body, so it can tap the turn it
// arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6f858405-dfe4-4ca2-b2e3-80ce8cda5522",
		Name:            "Vulshok Sorcerer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
