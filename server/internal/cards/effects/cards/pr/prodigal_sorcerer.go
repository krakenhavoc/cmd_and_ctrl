package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prodigal Sorcerer — Creature — Human Wizard Sorcerer {2}{U}, 1/1 (Secret
// Lair Drop, #5462):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Rootwater Hunter's {T} ping: it can't be activated while the Sorcerer is
// summoning sick (CR 302.6), and a target gone in response takes nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe",
		Name:         "Prodigal Sorcerer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
