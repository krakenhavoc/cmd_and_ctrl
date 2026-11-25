package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blossom Dryad — Creature — Dryad {2}{G}, 2/2 (Ixalan, S82 older sets):
//
//	"{T}: Untap target land."
//
// Voyaging Satyr's ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53929852-6359-4e7a-8f2c-a5954c76ffcc",
		Name:         "Blossom Dryad",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target land.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.B22UntapEachLegalTarget,
		}},
	})
}
