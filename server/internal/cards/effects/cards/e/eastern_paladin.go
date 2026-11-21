package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eastern Paladin — Creature — Phyrexian Zombie Knight {2}{B}{B}, 3/3
// (Urza's Saga):
//
//	"{B}{B}, {T}: Destroy target green creature."
//
// Exorcist's row with a green target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b445482a-9422-41c7-bedb-3b9f500d2f27",
		Name:         "Eastern Paladin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, {T}: Destroy target green creature.",
			Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target green creature", effects.OfColor("G")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
