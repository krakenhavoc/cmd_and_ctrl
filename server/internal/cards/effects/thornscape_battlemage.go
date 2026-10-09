package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thornscape Battlemage — Creature — Elf Wizard {2}{G}, 2/2:
//
//	"Kicker {R} and/or {W} (You may pay an additional {R} and/or {W}
//	 as you cast this spell.)
//	 When this creature enters, if it was kicked with its {R} kicker,
//	 it deals 2 damage to any target.
//	 When this creature enters, if it was kicked with its {W} kicker,
//	 destroy target artifact."
//
// Two kicker costs (CR 702.33b, #2153), each linked to one enters
// ability (CR 702.33f). ThisKickedWith reads which kicker was paid off
// the record the resolution carried onto the permanent (CR 400.7d), as
// an intervening if (CR 603.4): kicked with {R} alone, only the damage
// triggers; with both, both do, each with its own target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8d4d6806-cb01-49e3-91cc-fb0e7f7a8684",
		Name:          "Thornscape Battlemage",
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers("{R}", "{W}"),
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(On(game.EventETB, AllOf(Self, ThisKickedWith("{R}")),
				"Thornscape Battlemage — kicked with {R}, 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
			Targeting(On(game.EventETB, AllOf(Self, ThisKickedWith("{W}")),
				"Thornscape Battlemage — kicked with {W}, destroy target artifact",
				destroyChosenPermanent), TargetPermanent("target artifact", Artifact())),
		},
	})
}
