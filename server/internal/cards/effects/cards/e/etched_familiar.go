package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etched Familiar — Artifact Creature — Phyrexian Fox 3/2 {2}{B} (March of the Machine, S82):
//
//	"When this creature dies, each opponent loses 2 life and you gain 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2bc064e-4739-4eb1-ba44-1fd92b2829d7",
		Name:         "Etched Familiar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Etched Familiar — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
