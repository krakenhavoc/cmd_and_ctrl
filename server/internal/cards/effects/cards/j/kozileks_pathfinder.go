package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Pathfinder — Creature — Eldrazi {6}, 5/5 (Oath of the Gatewatch,
// #5545):
//
//	"{C}: Target creature can't block this creature this turn."
//
// Screeching Griffin's row paid with colorless mana: the one target can't
// block this Pathfinder this turn and may still block others.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de5a8edb-c41a-4f3a-9231-c54486ba7739",
		Name:         "Kozilek's Pathfinder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{C}: Target creature can't block this creature this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{C}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBlockThisCreature("Kozilek's Pathfinder — can't block it"),
		}},
	})
}
