package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Sparkmage — Creature — Human Shaman {2}{R}, 0/1 (Worldwake, #5582):
//
//	"Haste
//	 {T}: This creature deals 1 damage to any target."
//
// Anaba Shaman's ping without the mana, on a creature with haste so it
// can tap the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab4a25d8-e1d0-454e-9d06-f7960a0352bc",
		Name:            "Cunning Sparkmage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
