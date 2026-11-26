package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalastria Healer — Creature — Vampire Cleric Ally {1}{B}, 1/2 (Battle for
// Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control enters, each
//	 opponent loses 1 life and you gain 1 life."
//
// One trigger per entering Ally.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b70718c3-a959-402a-a715-7a3c9724a70a",
		Name:         "Kalastria Healer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB313Rally("Kalastria Healer — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
