package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cascading Cataracts — Land:
//
//	"Indestructible
//	 {T}: Add {C}.
//	 {5}, {T}: Add five mana in any combination of colors."
//
// Five INDEPENDENT any-colour slots (AnyCombinationOfColors), the
// shape Selvala uses, so the controller can answer {W}{U}{B}{R}{G} or
// five of one colour. The ability has no identity narrowing because the
// printed text names none. Indestructible is a printed keyword, as on
// Darksteel Citadel.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d98b4250-3492-4864-9c4c-42db09b3ccd4",
		Name:            "Cascading Cataracts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{5}"},
				Produced: effects.AnyCombinationOfColors(5),
				Label:    "{5}, {T}: Add five mana in any combination of colors",
			},
		},
	})
}
