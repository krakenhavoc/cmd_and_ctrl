package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Symbiote Spawn — Creature — Symbiote Villain {2}{B}, 3/2 (Marvel's
// Spider-Man, #4460):
//
//	"When this creature dies, each opponent loses 2 life and you gain 2
//	 life."
//
// "You" is the controller as it last existed (CR 603.10a), which the
// trigger item carries. Each opponent loses first, then you gain.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c3ff360-461a-44f3-a8f2-eb061b149cf5",
		Name:         "Symbiote Spawn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Symbiote Spawn — each opponent loses 2 life and you gain 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
