package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bear Trap — Artifact {1} (Duskmourn: House of Horror, #4466):
//
//	"Flash
//	 {3}, {T}, Sacrifice this artifact: It deals 3 damage to target
//	 creature."
//
// Mouser Foundry's sacrifice-for-damage row with a tap added to the
// cost. The sacrificed Trap deals the damage on its last-known
// information (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f7afb2c3-50ee-40f5-8b14-32f80c2bacd0",
		Name:            "Bear Trap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
