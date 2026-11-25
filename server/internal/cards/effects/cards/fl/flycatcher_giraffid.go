package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flycatcher Giraffid — Creature — Antelope Lizard {4}{G}, 3/5 (Ikoria: Lair of Behemoths, #5512):
//
//	"This creature enters with your choice of a reach counter or a vigilance
//	 counter on it."
//
// The controller is asked as it enters and the counter lands before anything
// else can act; a keyword counter gives its keyword (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0264b059-75cb-41e6-9e51-262ac22b4c60",
		Name:         "Flycatcher Giraffid",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.S82OldB113EntersWithChoiceOfCounters("Flycatcher Giraffid", []string{game.CounterReach}, []string{game.CounterVigilance}),
	})
}
