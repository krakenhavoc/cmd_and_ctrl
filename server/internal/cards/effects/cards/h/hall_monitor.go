package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hall Monitor — Creature — Lizard Shaman {R}, 1/1 (Strixhaven: School of
// Mages, #5502):
//
//	"Haste
//	 {1}{R}, {T}: Target creature can't block this turn."
//
// Haste is the printed keyword, so the {T} row can be used the turn it
// arrives; the row is Chainwhip Cyclops's with a tap added to the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "937a04c2-41be-4ff3-8992-6628a9ee9792",
		Name:            "Hall Monitor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}: Target creature can't block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Hall Monitor — target creature can't block this turn"),
		}},
	})
}
