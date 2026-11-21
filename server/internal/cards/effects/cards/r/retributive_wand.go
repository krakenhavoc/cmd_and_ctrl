package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b550a94d-3a92-416e-b651-510cdbdd05f1",
		Name:         "Retributive Wand",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: This artifact deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisDies("Retributive Wand — 5 damage to any target", effects.SourceDealsDamageToEachLegalTarget(5)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 5))),
		},
	})
}
