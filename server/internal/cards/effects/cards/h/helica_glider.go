package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helica Glider — Creature — Nightmare Squirrel {2}{W}, 2/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"This creature enters with your choice of a flying counter or a first strike counter on it."
//
// The controller is asked as it enters and the counter lands before anything
// else can act; a keyword counter gives its keyword (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bcdb107-3827-4280-9e5e-7f1e771ef0c5",
		Name:         "Helica Glider",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.S82OldB113EntersWithChoiceOfCounters("Helica Glider", []string{game.CounterFlying}, []string{game.CounterFirstStrike}),
	})
}
