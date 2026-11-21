package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twinshot Sniper — Artifact Creature — Goblin Archer {3}{R}, 2/3 (Neon
// Dynasty, #5493):
//
//	"Reach
//	 When this creature enters, it deals 2 damage to any target.
//	 Channel — {1}{R}, Discard this card: It deals 2 damage to any target."
//
// Both halves deal the damage with the card as the source, the channel from
// the hand (CR 207.2c) after its discard cost, so the source is read as it
// last was (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b72fd535-0051-4f59-9d99-75895f751006",
		Name:            "Twinshot Sniper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Twinshot Sniper — 2 damage to any target", effects.SourceDealsDamageToEachLegalTarget(2)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Channel — {1}{R}, Discard this card: It deals 2 damage to any target.",
			Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
