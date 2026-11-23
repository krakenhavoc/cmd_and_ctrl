package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Onyx Mage — Creature — Human Wizard {1}{B}, 2/1 (Magic 2012, #5573):
//
//	"{1}{B}: Target creature you control gains deathtouch until end of
//	 turn. (Any amount of damage it deals to a creature is enough to
//	 destroy it.)"
//
// Raccoon Rallier's row for deathtouch, at a mana cost and instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a398777-2fa5-4cac-adee-4582e59d410d",
		Name:         "Onyx Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Target creature you control gains deathtouch until end of turn.",
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Effect:  effects.FirstTargetGainsUntilEndOfTurn(0, "Onyx Mage — deathtouch until end of turn", "deathtouch"),
		}},
	})
}
