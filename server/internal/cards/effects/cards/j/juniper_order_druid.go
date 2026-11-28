package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Juniper Order Druid — Creature — Human Cleric Druid {2}{G}, 1/1 (Ice Age, S82 OLD #5636):
//
//	"{T}: Untap target land."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "943ab696-470b-488f-9567-c8cd89ec06e3",
		Name:         "Juniper Order Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target land.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
