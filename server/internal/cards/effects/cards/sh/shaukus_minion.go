package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shauku's Minion — Creature — Human Minion {1}{B}{R}, 2/2 (Mirage, #5633):
//
//	"{B}{R}, {T}: This creature deals 2 damage to target white creature."
//
// The target is judged as white through the layers, at announce and again as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cf1d7c3-6ebe-440b-88c8-33b7f6aa0d7c",
		Name:         "Shauku's Minion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}, {T}: This creature deals 2 damage to target white creature.",
			Cost:    effects.Plus(effects.ManaCost("{B}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target white creature", effects.OfColor("W")),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
