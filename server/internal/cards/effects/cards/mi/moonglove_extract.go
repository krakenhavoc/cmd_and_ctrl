package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonglove Extract — Artifact {3} (Duel Decks: Elspeth vs. Tezzeret, #5578):
//
//	"Sacrifice this artifact: It deals 2 damage to any target."
//
// The sacrifice is the whole cost; "it" is the Extract, so the damage comes
// from its last-known information (CR 608.2h) and a target that left in
// response leaves the ability to do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac2711b7-318d-478a-b69c-d44baff46f2d",
		Name:         "Moonglove Extract",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this artifact: It deals 2 damage to any target.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
