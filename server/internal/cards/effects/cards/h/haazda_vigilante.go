package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haazda Vigilante — Creature — Giant Soldier {4}{W}, 4/4 (Murders at
// Karlov Manor, #4469):
//
//	"Whenever this creature enters or attacks, put a +1/+1 counter on
//	 target creature you control with power 2 or less."
//
// Sun Titan's two conditions (WhenThisEntersOrAttacks). The power bound
// is the target's, judged as the trigger goes on the stack and again as
// it resolves (CR 608.2b): a creature pumped past 2 in response gets
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b707ef4-2c17-41e7-8dac-583791a98055",
		Name:         "Haazda Vigilante",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Haazda Vigilante — a +1/+1 counter on target creature you control with power 2 or less",
					effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control with power 2 or less", effects.YouControl(), effects.PowerLE(2))),
		},
	})
}
