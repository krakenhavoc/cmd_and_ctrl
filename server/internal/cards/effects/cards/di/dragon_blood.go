package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Blood — Artifact {3} (Urza's Saga):
//
//	"{3}, {T}: Put a +1/+1 counter on target creature."
//
// Tangle Tumbler's counter row on an artifact that is only that.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b751c1c6-195d-4023-9d0b-3c91d6b834d4",
		Name:         "Dragon Blood",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Put a +1/+1 counter on target creature.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PlusOneCounterOnTheTarget,
		}},
	})
}
