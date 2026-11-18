package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b72fd535-0051-4f59-9d99-75895f751006",
		Name:            "Twinshot Sniper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Twinshot Sniper — 2 damage to any target", sourceDealsDamageToEachLegalTarget(2)),
				TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {1}{R}, Discard this card: It deals 2 damage to any target.",
			Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{1}{R}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetAny(),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
