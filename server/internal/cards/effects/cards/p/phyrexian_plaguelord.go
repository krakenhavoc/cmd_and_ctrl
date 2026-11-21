package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Plaguelord — {3}{B}{B} Creature — Phyrexian Carrier 4/4 (Jumpstart 2022, #5948):
//
//	"{T}, Sacrifice this creature: Target creature gets -4/-4 until end of turn.
//	 Sacrifice a creature: Target creature gets -1/-1 until end of turn."
//
// Plagued Rusalka's -1/-1 outlet, and the same shape at -4/-4 with the
// source's own {T} and sacrifice as the cost. The second ability's
// sacrifice may be the Plaguelord itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aff9e844-9e03-490b-b44f-10d385738cc6",
		Name:         "Phyrexian Plaguelord",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}, Sacrifice this creature: Target creature gets -4/-4 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Phyrexian Plaguelord — -4/-4", -4, -4),
			},
			{
				Label:   "Sacrifice a creature: Target creature gets -1/-1 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
				Cost:    effects.SacrificeACreature(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Phyrexian Plaguelord — -1/-1", -1, -1),
			},
		},
	})
}
