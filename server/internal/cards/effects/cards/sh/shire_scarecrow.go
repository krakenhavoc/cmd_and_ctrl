package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shire Scarecrow — Artifact Creature — Scarecrow {2}, 0/3 (The Lord of the
// Rings: Tales of Middle-earth):
//
//	"Defender
//	 {1}: Add one mana of any color. Activate only once each turn."
//
// A mana ability (CR 605.1a: no target, adds mana) with a mana cost, Ramos's
// OncePerTurn bit for the limit (counted per object, CR 602.5b).
//
// No simplification.
func init() {
	const label = "{1}: Add one mana of any color. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "4b9f6a6e-c168-4732-b454-2272d56636da",
		Name:            "Shire Scarecrow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|U|B|R|G}",
			Label:       label,
			OncePerTurn: true,
		}},
	})
}
