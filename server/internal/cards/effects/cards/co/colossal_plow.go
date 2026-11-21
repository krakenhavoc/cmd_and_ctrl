package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4cdf524a-1a99-4143-823b-d3c5c867c580",
		Name:         "Colossal Plow",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Colossal Plow — add {W}{W}{W} and gain 3 life",
				effects.Do(effects.AddMana{
					Produced: "{W}{W}{W}",
					Riders:   []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
				}, effects.GainLife{Amount: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 6",
			Cost:    effects.CrewCost(6),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Colossal Plow"),
		}},
	})
}
