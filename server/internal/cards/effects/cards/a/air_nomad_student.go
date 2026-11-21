package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air Nomad Student — Creature — Human Monk {3}{W}, 2/2 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Flying
//	 At the beginning of your end step, if this creature didn't attack
//	 this turn, put a +1/+1 counter on it."
//
// An intervening "if" (CR 603.4): the turn's attack record is read as
// the end step begins and again as the trigger resolves. The record is
// per object, so a Student that attacked keeps the fact through the end
// step even if it has untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a468f6ab-c0a8-4295-8b01-7147d3f889b4",
		Name:            "Air Nomad Student",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.AvB27CondThisDidntAttack{},
				effects.AtYourEndStep("Air Nomad Student — put a +1/+1 counter on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
