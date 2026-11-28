package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liberated Dwarf — Creature — Dwarf {R}, 1/1 (Judgment):
//
//	"{R}, Sacrifice this creature: Target green creature gets +1/+0 and
//	 gains first strike until end of turn."
//
// Ridged Kusite's pump-and-grant on a sacrifice cost, with the target
// narrowed to green creatures (read through the layered view as the ability
// is announced). The Dwarf is gone as a cost, so it is not a target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5081f2d-ebe2-4ff3-a8ef-fe50f8ecde60",
		Name:         "Liberated Dwarf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Target green creature gets +1/+0 and gains first strike until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")), game.AnswerPump),
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target green creature", effects.OfColor("G")),
			Effect:  effects.PumpAndGrantFirstTarget(1, 0, "Liberated Dwarf — +1/+0 and first strike until end of turn", "first strike"),
		}},
	})
}
