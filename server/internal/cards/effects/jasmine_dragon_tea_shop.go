package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jasmine Dragon Tea Shop — Land (Avatar: The Last Airbender, #4459):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast an Ally
//	 spell or activate an ability of an Ally source.
//	 {5}, {T}: Create a 1/1 white Ally creature token."
//
// Avengers Tower's restricted any-colour row with Ally for Hero: the
// mana's spend restriction is a disjunction of "cast an Ally spell" and
// "activate an ability of an Ally source", read off the spell or the
// source as it is spent (CR 106.6). The token is the set's 1/1 white Ally.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9a24444-289f-473f-9985-8df275257555",
		Name:         "Jasmine Dragon Tea Shop",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			colorlessTapRow(),
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color (Ally spells and Ally abilities only).",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Ally")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Ally")},
				)},
			},
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Create a 1/1 white Ally creature token.",
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1}),
		}},
	})
}
