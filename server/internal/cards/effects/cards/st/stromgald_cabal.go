package cardsst

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Stromgald Cabal — Creature — Human Knight {1}{B}{B}, 2/2 (Ice Age, S82 OLD
// batch 850, #5636):
//
//	"{T}, Pay 1 life: Counter target white spell."
//
// Order of the Sacred Torch's counter, aimed at white spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8312964-5cf0-435a-8757-ea1bb67258a4",
		Name:         "Stromgald Cabal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay 1 life: Counter target white spell.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayLife(1)),
			Targets: effects.TargetSpell("target white spell", effects.OfColor("W")),
			Effect:  effects.CounterTheChosenSpell,
		}},
	})
}
