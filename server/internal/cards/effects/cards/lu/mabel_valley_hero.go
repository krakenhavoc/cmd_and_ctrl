package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mabel, Valley Hero — Legendary Creature — Mouse Soldier {1}{R}{W},
// 1/2:
//
//	"Whenever Mabel or another creature you control enters, put a +1/+1
//	 counter on target creature that entered this turn."
//
// The trigger is "a creature you control enters", Mabel included. Its
// target is ANY creature that entered the battlefield this turn,
// whoever controls it, and is chosen as the ability goes on the stack;
// the entering creature itself always qualifies, so the trigger is only
// dropped if that creature is already gone (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab56f7cf-9538-45b5-a657-1e2ebef1e92b",
		Name:         "Mabel, Valley Hero",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
					"Mabel, Valley Hero — a +1/+1 counter on target creature that entered this turn",
					effects.PutACounterOnTheTarget(game.CounterPlusOne)),
				effects.TargetCreature("target creature that entered this turn", effects.RfCreatureCEnteredThisTurn())),
		},
	})
}
