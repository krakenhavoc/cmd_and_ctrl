package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retributive Wand — Artifact {3} (Core Set 2020, #5518):
//
//	"{3}, {T}: This artifact deals 1 damage to any target.
//	 When this artifact is put into a graveyard from the battlefield, it
//	 deals 5 damage to any target."
//
// The leave trigger is the dies family's "put into a graveyard from the
// battlefield" (Chromatic Star's), and the Wand deals the damage as it
// last existed (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b550a94d-3a92-416e-b651-510cdbdd05f1",
		Name:         "Retributive Wand",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: This artifact deals 1 damage to any target.",
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetAny(),
			Effect:  sourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisDies("Retributive Wand — 5 damage to any target", sourceDealsDamageToEachLegalTarget(5)),
				TargetAny()), ForTargets(DamageToTarget(0, 5))),
		},
	})
}
