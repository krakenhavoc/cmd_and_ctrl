package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigardian Priest — Creature — Human Cleric {1}{W}, 1/2 (Eldritch Moon,
// #5541):
//
//	"{1}, {T}: Tap target non-Human creature."
//
// Avacynian Priest's ability, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2e22c9d-b0ec-42bb-a520-3c36df546ba0",
		Name:         "Sigardian Priest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target non-Human creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target non-Human creature", effects.Not(effects.HasSubtype("Human"))),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
