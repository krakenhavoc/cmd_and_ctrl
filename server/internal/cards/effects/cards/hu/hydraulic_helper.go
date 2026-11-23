package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hydraulic Helper — {1}{U} Artifact Creature — Robot 2/3 (Marvel Super
// Heroes, #4455):
//
//	"Defender
//	 {T}: Add {U}. This mana can't be spent to cast a nonartifact spell."
//
// A mana ability (CR 605.1a) whose mana carries the restriction: it pays
// for artifact spells and for abilities, and the payment path refuses it
// for any other spell (the same tag The Mightstone and Weakstone uses).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "503928cd-1abd-44cd-92c3-6dc19bb07f78",
		Name:            "Hydraulic Helper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{U}",
			Label:        "Add {U}. This mana can't be spent to cast a nonartifact spell.",
			Restrictions: []string{effects.ManaRestrictNotNonartifactSpell},
		}},
	})
}
