package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Apprentice — Creature — Human Wizard {1}{U}{U} 1/1 (Eighth Edition,
// S82 older sets batch 620, #5607):
//
//	"{T}, Sacrifice this creature: Counter target spell."
//
// Ertai, Wizard Adept's counter with a tap and a self-sacrifice cost, which
// summoning sickness does stop (it has a {T} cost, CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "863ad4b8-094b-46b3-b6cd-0640e1f01d6d",
		Name:         "Daring Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Counter target spell.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetSpell("target spell"),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
