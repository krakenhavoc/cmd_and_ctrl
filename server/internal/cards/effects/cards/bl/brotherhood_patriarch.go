package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brotherhood Patriarch — Creature — Human Assassin {3}{B}, 4/1
// (Assassin's Creed, #5466):
//
//	"When this creature dies, each opponent loses 2 life and you gain 2
//	 life."
//
// Symbiote Spawn's trigger: each opponent loses first, then you gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c766a86-9c70-465e-874b-9d9f3b4eb090",
		Name:         "Brotherhood Patriarch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Brotherhood Patriarch — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
