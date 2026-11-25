package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nocturnal Feeder — Creature — Vampire Rogue {2}{B}, 2/1 (Jumpstart,
// #5510):
//
//	"Flying
//	 When this creature dies, each opponent loses 2 life and you gain 2
//	 life."
//
// Brotherhood Patriarch's dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f71bbaf2-02ea-4fc0-9a39-49d6689b36cd",
		Name:            "Nocturnal Feeder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Nocturnal Feeder — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
