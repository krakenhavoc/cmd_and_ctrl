package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainwhip Cyclops — Creature — Cyclops Warrior {4}{R}, 4/4 (Ravnica:
// Clue Edition):
//
//	"{3}{R}: Target creature can't block this turn."
//
// Bola Warrior's body (targetCreatureCantBlockThisTurn) on a plain mana
// cost: the restriction ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0324f33-3f25-4374-affa-c7b5b54bc8d7",
		Name:         "Chainwhip Cyclops",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Target creature can't block this turn.",
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Effect:  effects.TargetCreatureCantBlockThisTurn("Chainwhip Cyclops — target creature can't block this turn"),
		}},
	})
}
