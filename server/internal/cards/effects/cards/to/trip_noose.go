package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trip Noose — Artifact {2} (Duel Decks: Elspeth vs. Tezzeret, #5578):
//
//	"{2}, {T}: Tap target creature."
//
// One activated ability with a creature target; the tap is applied to the
// first target still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "456f2991-ab64-41bc-9bd1-a1b1e7993ed6",
		Name:         "Trip Noose",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
