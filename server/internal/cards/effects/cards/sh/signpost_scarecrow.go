package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Signpost Scarecrow — Artifact Creature — Scarecrow {4}, 2/4 (Throne of Eldraine, #5516):
//
//	"Vigilance
//	 {2}: Add one mana of any color."
//
// Crossroads Candleguide's filter: a mana ability with a mana cost and no {T}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b6c6dab-0687-4297-b215-89e699d65681",
		Name:            "Signpost Scarecrow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color"),
		},
	})
}
