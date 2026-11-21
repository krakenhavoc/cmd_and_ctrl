package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Candy Trail — Artifact — Food Clue {1} (Wilds of Eldraine, #4471):
//
//	"When this artifact enters, scry 2.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life and draw a card."
//
// One printed activated ability that is both the Food and the Clue
// payoffs: the cost is paid in full ({2}, tap, sacrifice) and the effect
// gains 3 life, then draws a card. The Food Clue subtypes make it count as
// a Food for "sacrifice a Food" and as a Clue for "sacrifice a Clue".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c8e4f29-53e0-4242-8ef8-7a9476dce178",
		Name:         "Candy Trail",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Candy Trail — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: You gain 3 life and draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: effects.Do(
				effects.GainLife{Amount: 3},
				effects.DrawCards{N: 1},
			),
		}},
	})
}
