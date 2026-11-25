package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frillscare Mentor — Creature — Human Warrior {2}{R}, 3/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature enters, put a menace counter on target non-Human
//	 creature you control.
//	 {2}{R}, {T}: Put a +1/+1 counter on each creature you control with menace."
//
// The creature the counter goes on can be any non-Human creature, including
// one that already has menace. The activated ability reads menace through the
// layered view, so printed menace, a granted menace and a menace counter all
// count, and the set is read as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c9d217e-d7f8-43df-9bb6-58d29f93e2c4",
		Name:         "Frillscare Mentor",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB113KeywordCounterOnTargetNonHuman("Frillscare Mentor", game.CounterMenace)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Put a +1/+1 counter on each creature you control with menace.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Effect:  effects.S82OldB113CounterOnEachWithKeyword("menace"),
		}},
	})
}
