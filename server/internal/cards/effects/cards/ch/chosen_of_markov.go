package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chosen of Markov // Markov's Servant — Creature — Human {2}{B} 2/2 //
// Creature — Vampire 4/4:
//
//	Chosen of Markov: "{T}, Tap an untapped Vampire you control: Transform
//	 this creature."
//
// The second tap is a different permanent (CR 118.12 style "tap an untapped
// ..." cost): the Chosen itself is never the Vampire it taps, since it is a Human
// when the cost is paid. Markov's Servant has no abilities.
func init() {
	const oracle = "29328bcb-e991-4095-9439-f493c8fa3f82"
	const tapVampire = "an untapped Vampire you control"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Chosen of Markov",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Tap an untapped Vampire you control: Transform this creature.",
			Cost: game.AbilityCost{Tap: true, TapOthers: &game.TapOthersCost{
				Count:  1,
				Filter: effects.TargetPermanent(tapVampire, effects.HasSubtype("Vampire")),
				Label:  tapVampire,
			}},
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Markov's Servant",
		Completeness: effects.CompletenessFull,
	})
}
