package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archangel of Wrath — Creature — Angel {2}{W}{W}, 3/4:
//
//	"Kicker {B} and/or {R} (You may pay an additional {B} and/or {R}
//	 as you cast this spell.)
//	 Flying, lifelink
//	 When this creature enters, if it was kicked, it deals 2 damage to
//	 any target.
//	 When this creature enters, if it was kicked twice, it deals 2
//	 damage to any target."
//
// Two kicker costs (CR 702.33b, #2153), each its own once-only toggle.
// "Kicked twice" is both of them paid: CardKickedTimes counts the two
// together (CR 702.33d), so a doubly kicked Archangel puts both
// triggers on the stack and each picks its own target. Both clauses are
// intervening ifs (CR 603.4) read off the record the resolution carried
// onto the permanent (CR 400.7d), so a reanimated Archangel triggers
// neither. The Archangel deals the damage, so its lifelink gains the
// life (CR 702.15b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "022e97af-2a3a-4e13-9b6b-d34fcc8cf168",
		Name:            "Archangel of Wrath",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		OptionalCosts:   Kickers("{B}", "{R}"),
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Archangel of Wrath — kicked, 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
			TriggerWithPurpose(Targeting(On(game.EventETB, AllOf(Self, ThisKickedAtLeast(2)),
				"Archangel of Wrath — kicked twice, 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
