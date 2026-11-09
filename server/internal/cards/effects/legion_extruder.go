package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Legion Extruder — Artifact {1}{R} (The Big Score, #4468):
//
//	"When this artifact enters, it deals 2 damage to any target.
//	 {2}, {T}, Sacrifice another artifact: Create a 3/3 colorless Golem
//	 artifact creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c77dc11e-d95d-4e7e-bb24-d971c179e9b0",
		Name:         "Legion Extruder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Legion Extruder — 2 damage to any target",
				sourceDealsDamageToEachLegalTarget(2)), TargetAny()), ForTargets(DamageToTarget(0, 2))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice another artifact: Create a 3/3 colorless Golem artifact creature token.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeAnotherN(1, "another artifact", Artifact())),
			Purpose: game.Purpose{Tokens: 1},
			Effect:  Do(CreateToken{Template: b35GolemToken(), N: 1}),
		}},
	})
}
