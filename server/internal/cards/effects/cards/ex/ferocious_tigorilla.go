package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ferocious Tigorilla — Creature — Cat Ape {3}{R}, 4/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"This creature enters with your choice of a trample counter or a menace
//	 counter on it."
//
// The controller is asked as it enters and the counter lands before anything
// else can act; a keyword counter gives its keyword (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a7fef1e-3f11-4350-86da-6f1315fedc53",
		Name:         "Ferocious Tigorilla",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.S82OldB113EntersWithChoiceOfCounters("Ferocious Tigorilla", []string{game.CounterTrample}, []string{game.CounterMenace}),
	})
}
