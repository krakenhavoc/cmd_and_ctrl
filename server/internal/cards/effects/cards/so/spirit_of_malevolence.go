package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit of Malevolence — Creature — Spirit {1}{B}, 2/1 (Core Set 2021, #5511):
//
//	"When this creature dies, each opponent loses 1 life and you gain 1
//	 life."
//
// Etched Familiar's dies drain at 1. The gain does not depend on how much
// was lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15323e45-8260-43dc-b72b-efd661a58d75",
		Name:         "Spirit of Malevolence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Spirit of Malevolence — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
