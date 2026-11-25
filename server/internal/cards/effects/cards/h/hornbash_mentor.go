package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hornbash Mentor — Creature — Human Warrior {2}{G}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, put a trample counter on target non-Human
//	 creature you control.
//	 {2}{G}, {T}: Put a +1/+1 counter on each creature you control with
//	 trample."
//
// Trample is read through the layered view, so printed trample, a granted
// trample and a trample counter all count, and the set is read as the ability
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec3031ff-93db-48e9-9e09-335786cb85c2",
		Name:         "Hornbash Mentor",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB113KeywordCounterOnTargetNonHuman("Hornbash Mentor", game.CounterTrample)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}: Put a +1/+1 counter on each creature you control with trample.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
			Effect:  effects.S82OldB113CounterOnEachWithKeyword("trample"),
		}},
	})
}
