package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainbreaker — Artifact Creature — Scarecrow {2}, 3/3 (Shadowmoor, S82 OLD
// b532, #5591):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {3}, {T}: Remove a -1/-1 counter from target creature."
//
// The counters come in with the creature (CR 614.1c), so it is a 1/1 as it
// arrives. Its own ability may take them off itself or any creature, the counter
// through the effect-side adder at -1. A target with no -1/-1 counter loses
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02765930-e15e-4cc1-ac0d-040334a9c5bc",
		Name:         "Chainbreaker",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB532MinusOneCountersAtEntry("Chainbreaker", 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Remove a -1/-1 counter from target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB532RemoveAMinusOneCounterFromTarget,
		}},
	})
}
