package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grasping Dunes — Land — Desert (Amonkhet, #5535):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Put a -1/-1 counter on target creature.
//	 Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "47d16c11-3033-44f3-9a12-2daf3453cc5b",
		Name:          "Grasping Dunes",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}, Sacrifice this land: Put a -1/-1 counter on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets:      effects.Harmful(effects.TargetCreature("target creature")),
			SorcerySpeed: true,
			Effect:       effects.PutACounterOnTheTarget(game.CounterMinusOne),
		}},
	})
}
