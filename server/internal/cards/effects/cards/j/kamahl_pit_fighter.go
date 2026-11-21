package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kamahl, Pit Fighter — Legendary Creature — Human Barbarian {4}{R}{R},
// 6/1 (Odyssey, #5613):
//
//	"Haste (This creature can attack and {T} as soon as it comes under
//	 your control.)
//	 {T}: Kamahl deals 3 damage to any target."
//
// Zuran Spellcaster's pinger at 3. Haste is printed (CR 702.10), so the
// {T} row works the turn Kamahl arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "153bb46f-ee8a-4da3-aab4-ae884f6ff403",
		Name:            "Kamahl, Pit Fighter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Kamahl deals 3 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
