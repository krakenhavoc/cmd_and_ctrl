package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoldering Werewolf // Erupting Dreadwolf — {2}{R}{R} Creature —
// Werewolf Horror 3/2 // Creature — Eldrazi Werewolf 6/4 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	Front: "When this creature enters, it deals 1 damage to each of up to
//	        two target creatures.
//	        {4}{R}{R}: Transform this creature."
//	Back:  "Whenever this creature attacks, it deals 2 damage to any
//	        target."
//
// The enters trigger takes up to two creature targets as it goes on the
// stack and deals the damage to each one still legal as it resolves
// (CR 608.2b); the attack trigger aims at any target. The transform is the
// permanent turning over in place at instant speed.
//
// No simplification.
func init() {
	const oracle = "372a01f7-6e58-4424-9405-0ee4acbb9346"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Smoldering Werewolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Smoldering Werewolf — 1 damage to each of up to two target creatures",
				effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2)),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}{R}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{4}{R}{R}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Erupting Dreadwolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WheneverThisAttacks("Erupting Dreadwolf — 2 damage to any target",
				effects.DamageToFirstTarget(2)),
				effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
