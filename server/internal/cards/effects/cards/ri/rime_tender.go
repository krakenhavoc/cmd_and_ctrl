package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rime Tender — Snow Creature — Human Druid {1}{G}, 2/2 (Modern Horizons):
//
//	"{T}: Untap another target snow permanent."
//
// "Another" is object identity, never a name (it can't untap itself), and the
// target must have the snow supertype as it is now, judged as the ability is
// activated and again as it resolves (CR 608.2b). The {T} is a cost, so the
// Tender needs to be free of summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1141a12a-2ebd-4e7b-8c8c-c7a1f160244c",
		Name:         "Rime Tender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap another target snow permanent.",
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetPermanent("another target snow permanent", effects.Snow())),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
