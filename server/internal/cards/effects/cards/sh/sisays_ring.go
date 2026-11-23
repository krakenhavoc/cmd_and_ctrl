package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sisay's Ring — Artifact {4} (Visions, #5631):
//
//	"{T}: Add {C}{C}."
//
// One mana ability (CR 605.1a) that adds two colorless mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f5822ae-651f-410e-9316-5522eeb52d72",
		Name:         "Sisay's Ring",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "{T}: Add {C}{C}.",
		}},
	})
}
