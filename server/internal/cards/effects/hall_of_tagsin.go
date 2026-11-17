package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hall of Tagsin — Land (The Brothers' War, #5482):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {4}, {T}: Create a tapped Powerstone token."
//
// Capital City's two mana abilities, plus an activated ability (it
// makes a token, so it is not a mana ability, CR 605.1a) that taps the
// land and makes a tapped Powerstone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd809587-773c-47ec-8676-cefef6d1e38f",
		Name:         "Hall of Tagsin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Create a tapped Powerstone token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1}),
		}},
	})
}
