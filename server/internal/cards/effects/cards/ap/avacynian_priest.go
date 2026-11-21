package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avacynian Priest — Creature — Human Cleric {1}{W}, 1/2 (Innistrad
// Remastered):
//
//	"{1}, {T}: Tap target non-Human creature."
//
// "Non-Human" is read through the layered view of the target's creature
// types.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df7e296d-96de-4280-a4d8-49189544b9e3",
		Name:         "Avacynian Priest",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target non-Human creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target non-Human creature", effects.Not(effects.HasSubtype("Human"))),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
