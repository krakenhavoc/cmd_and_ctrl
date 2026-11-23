package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nebula Dragon — Creature — Dragon {6}{R}, 4/4 (Edge of Eternities,
// #4461):
//
//	"Flying
//	 When this creature enters, it deals 3 damage to any target."
//
// Ghitu Slinger's entry trigger with 3 in place of 2: the Dragon is the
// source of the damage, and a target gone by resolution takes none
// (sourceDealsDamageToEachLegalTarget, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "724b4a59-9c69-4f2b-a436-72183c62f1ff",
		Name:            "Nebula Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Nebula Dragon — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)), effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
