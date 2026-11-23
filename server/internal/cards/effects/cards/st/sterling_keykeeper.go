package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sterling Keykeeper — Creature — Human Mercenary {1}{W}, 2/2 (Outlaws
// of Thunder Junction, #4468):
//
//	"{2}, {T}: Tap target non-Mount creature."
//
// Frostbridge Guard's tap ability at {2} plus {T}, its target narrowed
// to creatures without the Mount subtype. The {T} in the cost means a
// summoning-sick Keykeeper can't activate it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f893d3d6-efef-4394-8e15-e01deed72b4f",
		Name:         "Sterling Keykeeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Tap target non-Mount creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target non-Mount creature", effects.Not(effects.Subtype("Mount"))),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
