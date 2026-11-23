package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Spawn — Creature — Vampire {2}{B}, 2/3 (Foundations, #4465):
//
//	"When this creature enters, each opponent loses 2 life and you gain
//	 2 life."
//
// Skirmish Rhino's drain on a black body: every opponent loses 2 life,
// then you gain 2, a fixed amount however many opponents there are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33cee4b4-3d80-46f1-9223-49ed785891a1",
		Name:         "Vampire Spawn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vampire Spawn — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
