package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Springjaw Trap — Artifact {1} (Ikoria: Lair of Behemoths, #5512):
//
//	"Flash
//	 {4}, {T}, Sacrifice this artifact: It deals 3 damage to any target."
//
// The damage comes from the Trap even though it was sacrificed as a cost, read
// from last-known information (CR 608.2h); an illegal target does nothing
// (CR 608.2b). An artifact that is not a creature has no summoning sickness,
// so the {T} can be paid the turn it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a635b73-692f-4345-9628-cb4f9f0cddad",
		Name:            "Springjaw Trap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: It deals 3 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
