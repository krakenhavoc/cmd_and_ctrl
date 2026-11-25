package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boot Nipper — Creature — Beast {1}{B}, 2/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"This creature enters with your choice of a deathtouch counter or a lifelink counter on it."
//
// The controller is asked as it enters and the counter lands before anything
// else can act; a keyword counter gives its keyword (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6337404-2978-47a8-a5ae-4a70add2df69",
		Name:         "Boot Nipper",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.S82OldB113EntersWithChoiceOfCounters("Boot Nipper", []string{game.CounterDeathtouch}, []string{game.CounterLifelink}),
	})
}
