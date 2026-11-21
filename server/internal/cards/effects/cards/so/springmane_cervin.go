package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Springmane Cervin — Creature — Elk {2}{G}, 3/2 (Strixhaven: School of
// Mages, #5502):
//
//	"When this creature enters, you gain 2 life."
//
// An enters trigger on the stack over a plain life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e964f31-c764-4af5-8c27-a46fd721fab4",
		Name:         "Springmane Cervin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Springmane Cervin — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
