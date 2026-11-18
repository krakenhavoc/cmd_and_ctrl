package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colossal Plow — Artifact — Vehicle {2}, 6/3:
//
//	"Whenever this Vehicle attacks, add {W}{W}{W} and you gain 3 life.
//	 Until end of turn, you don't lose this mana as steps and phases end.
//	 Crew 6"
//
// Brazen Collector's attack trigger with {W}{W}{W}: the mana carries the
// keep-until-end-of-turn rider (KeepManaUntilEndOfTurn), then the life is
// gained. Crew 6 is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4cdf524a-1a99-4143-823b-d3c5c867c580",
		Name:         "Colossal Plow",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Colossal Plow — add {W}{W}{W} and gain 3 life",
				Do(AddMana{
					Produced: "{W}{W}{W}",
					Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
				}, GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 6",
			Cost:    CrewCost(6),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Colossal Plow"),
		}},
	})
}
