package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Watcher of the Roost — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Bird Soldier {2}{W}, 2/1: Flying. Morph—Reveal a white card in your
//	hand. When this creature is turned face up, you gain 2 life.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "92cf18ab-dd7b-4bc8-8b85-a81b8cb5192c",
		Name:             "Watcher of the Roost",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.MorphReveal("W")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Watcher of the Roost — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
