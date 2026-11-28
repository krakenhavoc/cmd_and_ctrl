package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fyndhorn Brownie — Creature — Ouphe {2}{G}, 1/1 (Ice Age, S82 OLD #5636):
//
//	"{2}{G}, {T}: Untap target creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d1f8073-4da2-4abc-abb4-2053c4a40bbf",
		Name:         "Fyndhorn Brownie",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}: Untap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
