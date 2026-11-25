package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keensight Mentor — Creature — Human Cleric {2}{W}, 1/4 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, put a vigilance counter on target non-Human
//	 creature you control.
//	 {1}{W}, {T}: Put a +1/+1 counter on each creature you control with
//	 vigilance."
//
// Vigilance is read through the layered view, so printed vigilance, a granted
// vigilance and a vigilance counter all count, and the set is read as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "500a17a0-8332-414c-9c3a-7157a7800887",
		Name:         "Keensight Mentor",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB113KeywordCounterOnTargetNonHuman("Keensight Mentor", game.CounterVigilance)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Put a +1/+1 counter on each creature you control with vigilance.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Effect:  effects.S82OldB113CounterOnEachWithKeyword("vigilance"),
		}},
	})
}
