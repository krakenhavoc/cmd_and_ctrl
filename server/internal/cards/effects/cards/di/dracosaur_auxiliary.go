package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dracosaur Auxiliary — Creature — Dinosaur Dragon Mount {4}{R}{R}:
//
//	"Flying, haste
//	 Whenever this creature attacks while saddled, it deals 2 damage to
//	 any target.
//	 Saddle 3"
//
// The damage is the Dracosaur's own (sourceDealsDamageToEachLegalTarget),
// so lifelink, deathtouch and damage prevention read the right source.
//
// No simplification.
func init() {
	attack := effects.AttacksWhileSaddled("Dracosaur Auxiliary — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2))
	attack.Targets = effects.TargetAny()
	attack = effects.TriggerWithPurpose(attack, effects.ForTargets(effects.DamageToTarget(0, 2)))
	effects.Register(effects.Spec{
		OracleID:        "457e3218-b2ef-4280-931a-5320987b9657",
		Name:            "Dracosaur Auxiliary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(3)},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
