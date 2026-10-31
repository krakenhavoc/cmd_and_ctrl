package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:                   "d6f36172-d3ef-49e9-a55f-ce2983bb2439",
		Name:                       "Royal Talon Fighter Jet",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Royal Talon Fighter Jet — create a 1/1 white Soldier for each +1/+1 counter on it",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 white Soldier"),
					Name:     "the +1/+1 counters on it",
					Count: func(ctx *Context) int {
						self, ok := ctx.SourcePermanent()
						if !ok {
							return 0
						}
						return self.Counters[game.CounterPlusOne]
					},
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Royal Talon Fighter Jet"),
		}},
	})
}
