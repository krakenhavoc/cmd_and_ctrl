package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duct Crawler — Creature — Insect {R}, 1/1 (Stronghold, S82 OLD batch 746,
// #5627):
//
//	"{1}{R}: Target creature can't block this creature this turn."
//
// Spin Engine's row at {1}{R}: the record bars only blocks of this creature,
// and the target is read again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d47a7aa5-3b64-4ff8-bb10-e46a34d1993c",
		Name:         "Duct Crawler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Target creature can't block this creature this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{1}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBlockThisCreature("Duct Crawler — can't block it"),
		}},
	})
}
