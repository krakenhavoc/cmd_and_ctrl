package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voldaren Estate — Land (Commander, #5368):
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add one mana of any color. Spend this mana only to
//	 cast a Vampire spell.
//	 {5}, {T}: Create a Blood token. This ability costs {1} less to
//	 activate for each Vampire you control."
//
// The two mana abilities are Bucolic Ranch's pair with the life as part
// of the second's cost (CR 119.4, Mana Confluence): the restricted mana
// pays only a cast of a spell with the Vampire subtype (CR 106.6). The
// Blood ability's discount is its own cost clause (Diary of Dreams'
// shape), counting the Vampire creatures you control as it is activated,
// through the layers; the reduction never goes below the coloured and
// {T} parts, and generic mana can't drop below zero (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb0c0426-f1a6-4e52-9242-627786d3119a",
		Name:         "Voldaren Estate",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			{
				Cost:         ManaAbilityCost{Tap: true, Life: 1},
				Produced:     "{W|U|B|R|G}",
				Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Vampire")},
				Label:        "{T}, Pay 1 life: Add one mana of any color. Spend this mana only to cast a Vampire spell.",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Create a Blood token. This ability costs {1} less to activate for each Vampire you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(PermanentsYouControl(And(Creature(), HasSubtype("Vampire"))),
					"This ability costs {1} less to activate for each Vampire you control."),
			},
			Effect: Do(CreateToken{Template: BloodToken(), N: 1}),
		}},
	})
}
