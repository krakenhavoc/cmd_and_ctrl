package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Javelineers — Creature — Human Soldier {W}, 1/1 (Dominaria
// Remastered, #5948):
//
//	"This creature enters with a javelin counter on it.
//	 {T}, Remove a javelin counter from this creature: It deals 1 damage to
//	 any target."
//
// The counter is an entry replacement (CR 614.1c); removing it is a cost,
// paid with the tap, so the ability can be activated only while the counter
// is there. The damage is dealt by the creature (last-known information if
// it has gone).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26112ba7-2859-4af3-93c3-e62c31177216",
		Name:         "Icatian Javelineers",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("javelin", 1, "Icatian Javelineers: enters with a javelin counter"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a javelin counter from this creature: It deals 1 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis("javelin", 1)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
