package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tattered Mummy — Creature — Zombie Jackal {1}{B}, 1/2 (Game Night 2019,
// #5515):
//
//	"When this creature dies, each opponent loses 2 life."
//
// Brotherhood Patriarch's dies trigger without the life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "066e7379-d439-41cb-a8c9-60964040b110",
		Name:         "Tattered Mummy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tattered Mummy — each opponent loses 2 life",
				effects.Do(effects.EachOpponentLosesLife{N: 2})),
		},
	})
}
