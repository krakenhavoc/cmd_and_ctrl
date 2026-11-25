package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duskfang Mentor — Creature — Human Cleric {2}{B}, 1/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, put a lifelink counter on target non-Human
//	 creature you control.
//	 {1}{B}, {T}: Put a +1/+1 counter on each creature you control with
//	 lifelink."
//
// Lifelink is read through the layered view, so printed lifelink, a granted
// lifelink and a lifelink counter all count, and the set is read as the ability
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8b770c2-0406-4a2d-95d9-e14bb5d3bdc6",
		Name:         "Duskfang Mentor",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB114KeywordCounterOnTargetNonHuman("Duskfang Mentor", game.CounterLifelink)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Put a +1/+1 counter on each creature you control with lifelink.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Effect:  effects.S82OldB114CounterOnEachWithKeyword("lifelink"),
		}},
	})
}
