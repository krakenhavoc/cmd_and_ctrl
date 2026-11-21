package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tireless Missionaries — Creature — Human Cleric {4}{W}, 2/3 (Magic 2011,
// #5579):
//
//	"When this creature enters, you gain 3 life."
//
// Venerable Monk's enters trigger at 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a730e699-834a-4aac-9b3b-29d6917fc9b3",
		Name:         "Tireless Missionaries",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Tireless Missionaries — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
