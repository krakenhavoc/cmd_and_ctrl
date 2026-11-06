package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Melkotian Adjudicator — Creature — Melkotian Advisor {2}{U}, 2/1
// (Star Trek, #4973):
//
//	"Flying, vigilance
//	 Whenever you face a dilemma, put a +1/+1 counter on this creature.
//	 (You face a dilemma as you choose one or more modes for a spell or
//	 ability.)"
//
// "Face a dilemma" is the engine's EventFacedDilemma (batch 7, Jett
// Reno): its controller chose one or more modes for a spell or ability
// (CR 700.2a, 700.2b). The counter is a Do step, so the row declares its
// footprint. A creature that has left the battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1fdcdb1c-a140-40ee-b711-5ea539eb1084",
		Name:            "Melkotian Adjudicator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventFacedDilemma, YouFacedADilemma,
				"Melkotian Adjudicator — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
