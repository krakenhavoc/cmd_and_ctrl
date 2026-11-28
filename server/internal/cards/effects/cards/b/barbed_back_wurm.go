package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbed-Back Wurm — Creature — Wurm {4}{B}, 4/3 (Mirage, S82 OLD b797, #5633):
//
//	"{B}: Target green creature blocking this creature gets -1/-1 until end of
//	 turn."
//
// The target clause is Flowstone Salamander's: a blocker of this creature,
// judged at announce and at resolution, here also green. A creature that
// stopped blocking, or stopped being green, is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d60cbdf0-f20f-4306-be6b-eed90d2a2f77",
		Name:         "Barbed-Back Wurm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Target green creature blocking this creature gets -1/-1 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Cost:    effects.ManaCost("{B}"),
			Targets: effects.BlockingSource(effects.TargetCreature("target green creature blocking it", effects.OfColor("G"))),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Barbed-Back Wurm", -1, -1),
		}},
	})
}
