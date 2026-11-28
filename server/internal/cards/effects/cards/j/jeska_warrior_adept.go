package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeska, Warrior Adept — Legendary Creature — Human Barbarian Warrior
// {2}{R}{R}, 3/1 (Judgment, #5611):
//
//	"First strike, haste
//	 {T}: Jeska deals 1 damage to any target."
//
// Prodigal Sorcerer's ping with two printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3186fddd-23fd-440c-ad61-b4130e00f765",
		Name:            "Jeska, Warrior Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Jeska deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
