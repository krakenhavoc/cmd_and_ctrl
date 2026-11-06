package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "a468f6ab-c0a8-4295-8b01-7147d3f889b4",
		Name:            "Air Nomad Student",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(avB27CondThisDidntAttack{},
				AtYourEndStep("Air Nomad Student — put a +1/+1 counter on it",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
