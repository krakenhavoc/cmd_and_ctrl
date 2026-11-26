package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wingfold Pteron — Creature — Dinosaur {5}{U}, 3/6 (Ikoria: Lair of Behemoths, #5512):
//
//	"This creature enters with your choice of a flying counter or a hexproof counter on it."
//
// The controller is asked as it enters and the counter lands before anything
// else can act; a keyword counter gives its keyword (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7db3bd9-ef67-4d64-a503-08a3c03404b2",
		Name:         "Wingfold Pteron",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.S82OldB113EntersWithChoiceOfCounters("Wingfold Pteron", []string{game.CounterFlying}, []string{game.CounterHexproof}),
	})
}
