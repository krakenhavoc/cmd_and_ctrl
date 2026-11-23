package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tangled Florahedron // Tangled Vale — Modal DFC. Front: Creature —
// Elemental {1}{G}, 1/1: "{T}: Add {G}." Back: Land (mdfc_lands.go).
//
// An ordinary mana creature: the tap ability is a mana ability, subject
// to summoning sickness like any {T} cost. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53542c79-a62a-4d6a-97db-5296e9c68302",
		Name:         "Tangled Florahedron",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
