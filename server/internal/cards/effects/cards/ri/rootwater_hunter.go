package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootwater Hunter — Creature — Merfolk {2}{U}, 1/1 (Tempest, #5628):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Razorfin Hunter's text in mono-blue: the same {T} row, which can't be
// activated while the Hunter is summoning sick, and the shared
// sourceDealsDamageToEachLegalTarget body, so a target gone in response
// takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e42ed82-8b56-46ad-a1ca-e29689a9d030",
		Name:         "Rootwater Hunter",
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
