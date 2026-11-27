package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Civilized Scholar // Homicidal Brute — Creature — Human Advisor {2}{U},
// 0/1 // Creature — Human Mutant, 5/1 (Innistrad, S82 older sets batch 449,
// #5572):
//
//	"{T}: Draw a card, then discard a card. If a creature card is discarded
//	 this way, untap this creature, then transform it."
//	// "At the beginning of your end step, if this creature didn't attack
//	 this turn, tap this creature, then transform it."
//
// The loot is a draw whose discard waits for the draw to settle; the
// discard's answer looks at what was really discarded (CR 701.9) and turns
// the Scholar over only if it is still the permanent that was activated.
// The Brute's trigger is intervening-if (CR 603.4): it checks again as it
// resolves, then taps the Brute and transforms it.
//
// No simplification.
func init() {
	const oracle = "8662c9be-2244-4ea3-bf80-31a117f40a59"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Civilized Scholar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card. If a creature card is discarded this way, untap this creature, then transform it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.S82OldB449ScholarLootAndTurn,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Homicidal Brute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.AvB27CondThisDidntAttack{},
				effects.AtYourEndStep("Homicidal Brute — tap this creature, then transform it", effects.S82OldB449TapThenTransform)),
		},
	})
}
