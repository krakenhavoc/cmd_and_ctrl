package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Royal Talon Fighter Jet — {X}{W}{W} Artifact — Vehicle 1/1 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 This Vehicle enters with X +1/+1 counters on it.
//	 Whenever this Vehicle enters or attacks, create a number of 1/1 white
//	 Soldier creature tokens equal to the number of +1/+1 counters on it.
//	 Crew 2"
//
// The entry counters are Marketback Walker's XCounters clause, so they are on
// the permanent before its own enter trigger is put on the stack and the
// trigger counts a finished Vehicle. The count is read as the trigger resolves
// off the Vehicle, or off its last-known information if it has left
// (CR 608.2h), so a counter removed in response makes fewer Soldiers and a
// Jet that is gone still makes the number it had. Crew is Dependable
// Quinjet's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "d6f36172-d3ef-49e9-a55f-ce2983bb2439",
		Name:                       "Royal Talon Fighter Jet",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Royal Talon Fighter Jet — create a 1/1 white Soldier for each +1/+1 counter on it",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/1 white Soldier"),
					Name:     "the +1/+1 counters on it",
					Count:    effects.SocB9PlusOneCountersOnSource,
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Royal Talon Fighter Jet"),
		}},
	})
}
