package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Mystic — Creature — Elephant Cleric {3}{W}{W} (Darksteel, #5605):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tapper: the shared tapChosenPermanent body, the
// target re-checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3e22884-a9ca-4479-b400-df1c1d2fd153",
		Name:         "Loxodon Mystic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
